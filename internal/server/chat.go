package server

import (
	"context"
	"strings"

	"github.com/google/uuid"
	chatv1 "github.com/Suizer98/protobuf-ai-potato/gen/go/chat/v1"
	"github.com/Suizer98/protobuf-ai-potato/internal/llm"
	"github.com/Suizer98/protobuf-ai-potato/internal/session"
)

type ChatServer struct {
	chatv1.UnimplementedChatServiceServer
	provider llm.Provider
	sessions *session.Store
}

func NewChatServer(provider llm.Provider, sessions *session.Store) *ChatServer {
	return &ChatServer{
		provider: provider,
		sessions: sessions,
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

func ptr(value string) *string {
	return &value
}
