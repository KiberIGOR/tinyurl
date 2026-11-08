package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/KiberIGOR/tinyurl/internal/repository"
)

func BenchmarkShorten(b *testing.B) {
	m := repository.NewMemory()
	s := NewShortener(m, "http://localhost:8080/")
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orig := fmt.Sprintf("https://example.com/page/%d", i)
		if _, err := s.Shorten(ctx, orig, 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkResolve(b *testing.B) {
	m := repository.NewMemory()
	s := NewShortener(m, "http://localhost:8080/")
	ctx := context.Background()
	const short = "benchId01"
	if _, err := m.Save(ctx, short, "https://example.com/long", 1); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Resolve(ctx, short); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkResolveByUserID(b *testing.B) {
	m := repository.NewMemory()
	s := NewShortener(m, "http://localhost:8080/")
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		if _, err := m.Save(ctx, fmt.Sprintf("s%04d", i), fmt.Sprintf("https://ex.com/%d", i), 3); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.ResolveByUserID(ctx, 3); err != nil {
			b.Fatal(err)
		}
	}
}
