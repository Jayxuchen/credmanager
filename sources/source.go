package sources

import (
	"context"

	"github.com/jayxuchen/credmanager/types"
)

type CredentialSource interface {
	GetCredentials(ctx context.Context) (*types.Credential, error)
	Name() string
}
