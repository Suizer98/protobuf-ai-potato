package rag

import (
	"strings"
	"unicode/utf8"
)

const defaultChunkSize = 700
const defaultChunkOverlap = 100

func ChunkText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	paragraphs := splitParagraphs(text)
	var chunks []string
	var current strings.Builder

	flush := func() {
		part := strings.TrimSpace(current.String())
		if part != "" {
			chunks = append(chunks, part)
		}
		current.Reset()
	}

	for _, paragraph := range paragraphs {
		if utf8.RuneCountInString(paragraph) > defaultChunkSize {
			flush()
			chunks = append(chunks, splitLong(paragraph, defaultChunkSize, defaultChunkOverlap)...)
			continue
		}
		if current.Len() == 0 {
			current.WriteString(paragraph)
			continue
		}
		if utf8.RuneCountInString(current.String())+1+utf8.RuneCountInString(paragraph) <= defaultChunkSize {
			current.WriteByte('\n')
			current.WriteString(paragraph)
			continue
		}
		flush()
		current.WriteString(paragraph)
	}
	flush()
	return chunks
}

func splitParagraphs(text string) []string {
	raw := strings.Split(text, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return []string{text}
	}
	return out
}

func splitLong(text string, size, overlap int) []string {
	runes := []rune(text)
	if len(runes) <= size {
		return []string{text}
	}
	var out []string
	step := size - overlap
	if step < 1 {
		step = size
	}
	for start := 0; start < len(runes); start += step {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
		if end == len(runes) {
			break
		}
	}
	return out
}
