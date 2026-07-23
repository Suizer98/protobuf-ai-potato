package rag

import "testing"

func TestChunkText(t *testing.T) {
	chunks := ChunkText("alpha\n\nbeta\n\ngamma")
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
	joined := ""
	for _, c := range chunks {
		joined += c
	}
	if joined == "" {
		t.Fatal("empty joined chunks")
	}
}

func TestChunkTextEmpty(t *testing.T) {
	if got := ChunkText("   "); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}
