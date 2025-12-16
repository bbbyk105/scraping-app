package providers

import (
	"fmt"
)

type Manager struct {
	providers map[string]Provider
}

func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

func (m *Manager) Register(name string, provider Provider) {
	m.providers[name] = provider
}

func (m *Manager) Get(name string) (Provider, error) {
	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

func (m *Manager) List() []string {
	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

