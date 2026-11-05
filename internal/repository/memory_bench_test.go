package repository

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkMemorySave(b *testing.B) {
	m := NewMemory()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		short := fmt.Sprintf("id%08d", i)
		orig := fmt.Sprintf("https://example.com/path/%d", i)
		if _, err := m.Save(ctx, short, orig, 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryGet(b *testing.B) {
	m := NewMemory()
	ctx := context.Background()
	const short = "benchShort"
	if _, err := m.Save(ctx, short, "https://example.com/", 1); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := m.Get(ctx, short); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryGetURLsByUserID(b *testing.B) {
	m := NewMemory()
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		short := fmt.Sprintf("u%08d", i)
		if _, err := m.Save(ctx, short, "https://example.com/"+short, 42); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := m.GetURLsByUserID(ctx, 42); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryDeleteByShortURLs(b *testing.B) {
	m := NewMemory()
	ctx := context.Background()
	shorts := make([]string, 50)
	for i := range shorts {
		shorts[i] = fmt.Sprintf("d%08d", i)
		if _, err := m.Save(ctx, shorts[i], "https://example.com/", 7); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := m.DeleteByShortURLs(ctx, 7, shorts); err != nil {
			b.Fatal(err)
		}
		for _, s := range shorts {
			delete(m.deleted, s)
		}
	}
}
