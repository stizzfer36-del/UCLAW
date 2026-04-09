package providers

import (
	"context"
	"testing"
)

func TestMockProvider(t *testing.T) {
	p, err := New("mock", ""); if err != nil { t.Fatal(err) }
	value, err := p.Generate(context.Background(), "hello"); if err != nil { t.Fatal(err) }
	if value != "mock:hello" { t.Fatalf("unexpected provider response %q", value) }
}
