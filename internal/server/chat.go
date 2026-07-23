package server

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	chatv1 "github.com/Suizer98/protobuf-ai-potato/gen/go/chat/v1"
	"github.com/Suizer98/protobuf-ai-potato/internal/llm"
	"github.com/Suizer98/protobuf-ai-potato/internal/rag"
	"github.com/Suizer98/protobuf-ai-potato/internal/session"
)

type ChatServer struct {
	chatv1.UnimplementedChatServiceServer
	provider llm.Provider
	sessions *session.Store
	rag      *rag.Service
}

func NewChatServer(provider llm.Provider, sessions *session.Store, ragService *rag.Service) *ChatServer {
	return &ChatServer{
		provider: provider,
		sessions: sessions,
		rag:      ragService,
	}
}

func (s *ChatServer) Chat(req *chatv1.ChatRequest, stream chatv1.ChatService_ChatServer) error {
	ctx := stream.Context()
	sessionID := strings.TrimSpace(req.GetSessionId())
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return stream.Send(&chatv1.ChatChunk{
			SessionId: sessionID,
			Error:     ptr("message is required"),
			Done:      true,
		})
	}
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	s.sessions.GetOrCreate(sessionID, message)
	s.sessions.Append(sessionID, llm.Message{Role: llm.RoleUser, Content: message})

	history := s.sessions.Messages(sessionID)
	if s.rag != nil {
		contextText, err := s.rag.RetrieveContext(ctx, message)
		if err != nil {
			log.Printf("rag retrieve: %v", err)
		} else if contextText != "" {
			history = append([]llm.Message{{
				Role:    llm.RoleSystem,
				Content: contextText,
			}}, history...)
		}
	}

	chunks, err := s.provider.StreamChat(ctx, req.GetModel(), history)
	if err != nil {
		return stream.Send(&chatv1.ChatChunk{
			SessionId: sessionID,
			Error:     ptr(err.Error()),
			Done:      true,
		})
	}

	var assistant strings.Builder
	for chunk := range chunks {
		if chunk.Err != nil {
			return stream.Send(&chatv1.ChatChunk{
				SessionId: sessionID,
				Error:     ptr(chunk.Err.Error()),
				Done:      true,
			})
		}
		if chunk.Delta != "" {
			assistant.WriteString(chunk.Delta)
			if err := stream.Send(&chatv1.ChatChunk{
				SessionId: sessionID,
				Delta:     chunk.Delta,
			}); err != nil {
				return err
			}
		}
		if chunk.Done {
			break
		}
	}

	reply := assistant.String()
	if reply != "" {
		s.sessions.Append(sessionID, llm.Message{Role: llm.RoleAssistant, Content: reply})
	}

	return stream.Send(&chatv1.ChatChunk{
		SessionId: sessionID,
		Done:      true,
	})
}

func (s *ChatServer) ListSessions(ctx context.Context, _ *chatv1.ListSessionsRequest) (*chatv1.ListSessionsResponse, error) {
	items := s.sessions.List()
	out := &chatv1.ListSessionsResponse{
		Sessions: make([]*chatv1.Session, 0, len(items)),
	}
	for _, item := range items {
		out.Sessions = append(out.Sessions, &chatv1.Session{
			Id:            item.ID,
			Title:         item.Title,
			UpdatedAtUnix: item.UpdatedAt.Unix(),
		})
	}
	return out, nil
}

func (s *ChatServer) IngestDocument(ctx context.Context, req *chatv1.IngestDocumentRequest) (*chatv1.IngestDocumentResponse, error) {
	if s.rag == nil {
		return nil, status.Error(codes.FailedPrecondition, "rag is disabled (set DATABASE_URL)")
	}
	docID, chunkCount, err := s.rag.Ingest(ctx, req.GetTitle(), req.GetSource(), req.GetContent())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return &chatv1.IngestDocumentResponse{
		DocumentId: docID.String(),
		ChunkCount: int32(chunkCount),
	}, nil
}

func ptr(value string) *string {
	return &value
}
