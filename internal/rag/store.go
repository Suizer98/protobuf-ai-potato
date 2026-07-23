package rag

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	pool *pgxpool.Pool
}

type Hit struct {
	Content  string
	Distance float64
}

func OpenStore(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	store := &Store{pool: pool}
	if err := store.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, schemaSQL)
	return err
}

func (s *Store) InsertDocument(ctx context.Context, title, source string, contents []string, embeddings [][]float32) (uuid.UUID, error) {
	if len(contents) != len(embeddings) {
		return uuid.Nil, fmt.Errorf("contents/embeddings length mismatch")
	}
	docID := uuid.New()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO documents (id, title, source) VALUES ($1, $2, $3)`,
		docID, title, source,
	)
	if err != nil {
		return uuid.Nil, err
	}

	for i, content := range contents {
		_, err = tx.Exec(ctx,
			`INSERT INTO chunks (id, document_id, content, chunk_index, embedding)
			 VALUES ($1, $2, $3, $4, $5)`,
			uuid.New(), docID, content, i, pgvector.NewVector(embeddings[i]),
		)
		if err != nil {
			return uuid.Nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return docID, nil
}

func (s *Store) Search(ctx context.Context, query []float32, topK int) ([]Hit, error) {
	if topK <= 0 {
		topK = 4
	}
	rows, err := s.pool.Query(ctx, `
		SELECT content, embedding <=> $1 AS distance
		FROM chunks
		ORDER BY embedding <=> $1
		LIMIT $2
	`, pgvector.NewVector(query), topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		var hit Hit
		if err := rows.Scan(&hit.Content, &hit.Distance); err != nil {
			return nil, err
		}
		hits = append(hits, hit)
	}
	return hits, rows.Err()
}
