package credmanager

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jayxuchen/credmanager/sources"
	"github.com/jayxuchen/credmanager/types"
)

type CredentialManager struct {
	sources []sources.CredentialSource
	mu      sync.Mutex
}

func NewManager(srcs []sources.CredentialSource) *CredentialManager {
	return &CredentialManager{
		sources: srcs,
	}
}

// GetFirstValid attempts to fetch credentials from the first source that succeeds.
func (m *CredentialManager) GetFirstValid(ctx context.Context) (*types.Credential, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, src := range m.sources {
		creds, err := src.GetCredentials(ctx)
		if err == nil {
			return creds, nil
		}
		fmt.Printf("[credmanager] source %q failed: %v\n", src.Name(), err)
	}
	return nil, errors.New("no valid credential sources found")
}
