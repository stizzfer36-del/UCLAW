// Package agents defines the UCLAW agent runtime.
package agents

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Role enumerates the agent roles enforced by the policy engine.
type Role string

const (
	RoleDev     Role = "dev"
	RoleVerify  Role = "verify"
	RoleResearch Role = "research"
	RoleOps     Role = "ops"
	RoleLead    Role = "lead"
)

// Agent represents a running UCLAW agent.
type Agent struct {
	ID       string
	Name     string
	Role     Role
	Mission  string // mission_id
	Provider string // ollama | openai | anthropic | codex
	Model    string

	mu     sync.Mutex
	cancel context.CancelFunc
}

// Registry keeps track of live agents.
var Registry = map[string]*Agent{}
var mu sync.RWMutex

// Spawn registers and starts an agent.
func Spawn(a *Agent) error {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := Registry[a.ID]; exists {
		return fmt.Errorf("agent %s already running", a.ID)
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	Registry[a.ID] = a
	log.Printf("[agents] spawned %s (%s/%s) mission=%s", a.ID, a.Provider, a.Model, a.Mission)
	go a.run(ctx)
	return nil
}

// Kill stops an agent.
func Kill(id string) error {
	mu.Lock()
	defer mu.Unlock()
	a, ok := Registry[id]
	if !ok {
		return fmt.Errorf("agent %s not found", id)
	}
	a.cancel()
	delete(Registry, id)
	return nil
}

func (a *Agent) run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[agents] %s panicked: %v", a.ID, r)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			log.Printf("[agents] %s stopped", a.ID)
			return
		case <-time.After(30 * time.Second):
			// heartbeat — emit audit event
			_ = world.DB // keep compile dep
		}
	}
}

// List returns all registered agents.
func List() []*Agent {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Agent, 0, len(Registry))
	for _, a := range Registry {
		out = append(out, a)
	}
	return out
}
