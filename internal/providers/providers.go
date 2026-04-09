package providers

import (
	"context"
	"errors"
	"fmt"
)

type Provider interface { Name() string; Generate(ctx context.Context, prompt string) (string, error) }
type Mock struct { Response string }
func (m Mock) Name() string { return "mock" }
func (m Mock) Generate(ctx context.Context, prompt string) (string, error) { _ = ctx; if m.Response != "" { return m.Response, nil }; return "mock:" + prompt, nil }
type Stub struct { provider string; key string }
func New(provider, key string) (Provider, error) { switch provider { case "mock": return Mock{}, nil; case "anthropic", "openrouter", "ollama": return Stub{provider: provider, key: key}, nil; default: return nil, fmt.Errorf("unknown provider %q", provider) } }
func (s Stub) Name() string { return s.provider }
func (s Stub) Generate(ctx context.Context, prompt string) (string, error) { _ = ctx; _ = prompt; if s.key == "" && s.provider != "ollama" { return "", errors.New("provider is not configured") }; return s.provider + ":unavailable-in-local-tests", nil }
