package static

import (
	"context"
	"fmt"

	"github.com/jayxuchen/credmanager/types"
)

// RDSPostgresSource provides a single static credential for connecting to RDS Postgres.
type RDSPostgresSource struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func NewRDSPostgresSource(host string, port int, user, pass, db string) *RDSPostgresSource {
	return &RDSPostgresSource{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		Database: db,
	}
}

func (r *RDSPostgresSource) GetCredentials(ctx context.Context) (*types.Credential, error) {
	return &types.Credential{

		Key:    "rds_postgres",
		Value:  r.Password,
		Expiry: nil,
		Metadata: map[string]string{
			"host":     r.Host,
			"port":     fmt.Sprintf("%d", r.Port),
			"username": r.Username,
			"database": r.Database,
			"type":     "static",
			"service":  "rds_postgres",
		},
	}, nil
}

func (r *RDSPostgresSource) Name() string {
	return "static_rds_postgres"
}
