package dynamic

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"

	"github.com/jayxuchen/credmanager/types"
)

// DynamicRDSSource generates temporary RDS IAM authentication tokens
// Uses AWS SDK to generate DB auth tokens for RDS instances
// Suitable for EKS services with IAM roles
type DynamicRDSSource struct {
	Host     string
	Port     int
	Username string
	Database string
	Region   string
}

// NewDynamicRDSSource creates a new dynamic RDS credential source
// If parameters are empty, it will try to read from environment variables:
// DB_HOST, DB_PORT, DB_USER, DB_NAME, AWS_REGION
func NewDynamicRDSSource(host string, port int, user, db, region string) *DynamicRDSSource {
	// Use environment variables as fallback if parameters are empty
	if host == "" {
		host = os.Getenv("DB_HOST")
	}
	if port == 0 {
		if portStr := os.Getenv("DB_PORT"); portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				port = p
			}
		}
		if port == 0 {
			port = 5432 // Default PostgreSQL port
		}
	}
	if user == "" {
		user = os.Getenv("DB_USER")
	}
	if db == "" {
		db = os.Getenv("DB_NAME")
	}
	if region == "" {
		region = os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-west-2" // Default region
		}
	}

	return &DynamicRDSSource{
		Host:     host,
		Port:     port,
		Username: user,
		Database: db,
		Region:   region,
	}
}

// GetCredentials generates a new RDS IAM authentication token
// This method uses the AWS SDK to generate a temporary token that's valid for 15 minutes
// The service must have appropriate IAM permissions to generate DB auth tokens
func (d *DynamicRDSSource) GetCredentials(ctx context.Context) (*types.Credential, error) {
	// Create AWS session - will use IAM role from EKS service account
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Generate the DB auth token using AWS RDS utils
	// This creates a presigned URL that can be used as a password
	dbEndpoint := fmt.Sprintf("%s:%d", d.Host, d.Port)
	authToken, err := rdsutils.BuildAuthToken(dbEndpoint, d.Region, d.Username, sess.Config.Credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RDS auth token: %w", err)
	}

	// RDS auth tokens are valid for 15 minutes
	expiry := time.Now().Add(15 * time.Minute)

	return &types.Credential{
		Key:    "dynamic_rds_postgres",
		Value:  authToken,
		Expiry: &expiry,
		Metadata: map[string]string{
			"host":        d.Host,
			"port":        fmt.Sprintf("%d", d.Port),
			"username":    d.Username,
			"database":    d.Database,
			"region":      d.Region,
			"type":        "dynamic",
			"service":     "rds_postgres",
			"auth_method": "iam_token",
			"expires_at":  expiry.Format(time.RFC3339),
			"endpoint":    dbEndpoint,
		},
	}, nil
}

func (d *DynamicRDSSource) Name() string {
	return "dynamic_rds_postgres"
}

// IsTokenExpired checks if the current token has expired
func (d *DynamicRDSSource) IsTokenExpired(cred *types.Credential) bool {
	if cred.Expiry == nil {
		return false
	}
	return time.Now().After(*cred.Expiry)
}

// RefreshCredentials forces a refresh of the credentials
func (d *DynamicRDSSource) RefreshCredentials(ctx context.Context) (*types.Credential, error) {
	return d.GetCredentials(ctx)
}
