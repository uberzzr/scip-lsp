package indexer

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/uber/scip-lsp/src/ulsp/factory"
)

type cmdVal struct {
	token           string
	cancelFunc      context.CancelFunc
	needsReindexing bool
}

type pendingCmdStore struct {
	pendingCmds map[string]cmdVal
	mu          sync.Mutex
}

func (p *pendingCmdStore) setPendingCmd(key string, cancel context.CancelFunc) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingCmds == nil {
		p.pendingCmds = make(map[string]cmdVal)
	}

	token := factory.UUID().String()
	p.pendingCmds[key] = cmdVal{
		cancelFunc:      cancel,
		token:           token,
		needsReindexing: false,
	}

	return token
}

func (p *pendingCmdStore) getPendingCmd(key string) (cancelFunc context.CancelFunc, token string, entryExists bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingCmds == nil {
		return nil, "", false
	}

	val, ok := p.pendingCmds[key]
	return val.cancelFunc, val.token, ok
}

func (p *pendingCmdStore) deletePendingCmd(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingCmds == nil {
		return false
	}

	if _, ok := p.pendingCmds[key]; !ok {
		return false
	}

	delete(p.pendingCmds, key)
	return true
}

func (p *pendingCmdStore) cleanSession(uuid uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingCmds == nil {
		return
	}

	for key, val := range p.pendingCmds {
		if strings.HasPrefix(key, uuid.String()) {
			val.cancelFunc()
			delete(p.pendingCmds, key)
		}
	}
}

func (p *pendingCmdStore) getContainingKey(token string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingCmds == nil {
		return "", fmt.Errorf("no key found for token: %s", token)
	}

	for key, val := range p.pendingCmds {
		if val.token == token {
			return key, nil
		}
	}
	return "", fmt.Errorf("no key found for token: %s", token)
}

func (p *pendingCmdStore) markForReindexing(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingCmds == nil {
		return false
	}

	val, ok := p.pendingCmds[key]
	if !ok {
		return false
	}

	val.needsReindexing = true
	p.pendingCmds[key] = val
	return true
}

func (p *pendingCmdStore) needsReindexing(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pendingCmds == nil {
		return false
	}

	val, ok := p.pendingCmds[key]
	if !ok {
		return false
	}

	return val.needsReindexing
}
