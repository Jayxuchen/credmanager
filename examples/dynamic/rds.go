package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jayxuchen/credmanager"
	"github.com/jayxuchen/credmanager/sources"
	"github.com/jayxuchen/credmanager/sources/dynamic"
	"github.com/jayxuchen/credmanager/sources/static"
	"github.com/jayxuchen/credmanager/types"
)

func main() {
	fmt.Println("=== Dynamic RDS Credential Manager Example ===\n")
	fmt.Println("This example demonstrates AWS RDS IAM authentication token generation")
	fmt.Println("Suitable for EKS services with IAM roles attached\n")

	// Create a dynamic RDS credential source
	// It will use environment variables: DB_HOST, DB_PORT, DB_USER, DB_NAME, AWS_REGION
	// Or you can pass them explicitly
	dynamicRDSSource := dynamic.NewDynamicRDSSource(
		"", // Will use DB_HOST env var
		0,  // Will use DB_PORT env var (defaults to 5432)
		"", // Will use DB_USER env var
		"", // Will use DB_NAME env var
		"", // Will use AWS_REGION env var (defaults to us-west-2)
	)

	// Create a static RDS source as fallback
	staticRDSSource := static.NewRDSPostgresSource(
		"localhost",
		5432,
		"fallback_user",
		"fallback_password",
		"local_db",
	)

	// Create credential manager with multiple sources (dynamic first, static as fallback)
	sources := []sources.CredentialSource{dynamicRDSSource, staticRDSSource}
	manager := credmanager.NewManager(sources)

	// Demonstrate getting credentials multiple times
	for i := 1; i <= 3; i++ {
		fmt.Printf("--- Attempt %d ---\n", i)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Get credentials from the first valid source
		creds, err := manager.GetFirstValid(ctx)
		if err != nil {
			log.Printf("Failed to get credentials: %v", err)
			cancel()
			continue
		}

		// Display credential information
		fmt.Printf("✓ Source: %s\n", getSourceType(creds.Key))
		fmt.Printf("  Credential Key: %s\n", creds.Key)
		fmt.Printf("  Token Length: %d characters\n", len(creds.Value))

		if creds.Expiry != nil {
			timeUntilExpiry := time.Until(*creds.Expiry)
			fmt.Printf("  Expires: %s (in %v)\n",
				creds.Expiry.Format(time.RFC3339),
				timeUntilExpiry.Round(time.Second))
		} else {
			fmt.Println("  Expires: Never")
		}

		// Display metadata
		fmt.Println("  Metadata:")
		importantKeys := []string{"host", "port", "username", "database", "type", "auth_method", "region"}
		for _, key := range importantKeys {
			if value, exists := creds.Metadata[key]; exists {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}

		// Build connection string based on credential type
		connectionString := buildConnectionString(creds)
		fmt.Printf("  Connection: %s\n\n", connectionString)

		cancel()

		// Wait a bit before next attempt
		if i < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	// Demonstrate credential refresh scenario
	fmt.Println("--- Demonstrating Credential Refresh ---")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	creds, err := manager.GetFirstValid(ctx)
	if err != nil {
		log.Fatalf("Failed to get credentials for refresh demo: %v", err)
	}

	fmt.Printf("Initial token: %s...\n", creds.Value[:20])

	// Simulate waiting and refreshing
	fmt.Println("Simulating credential refresh...")
	time.Sleep(1 * time.Second)

	// Get new credentials (should generate new token for dynamic source)
	newCreds, err := manager.GetFirstValid(ctx)
	if err != nil {
		log.Fatalf("Failed to refresh credentials: %v", err)
	}

	fmt.Printf("Refreshed token: %s...\n", newCreds.Value[:20])

	if creds.Value != newCreds.Value && creds.Key == "dynamic_rds_postgres" {
		fmt.Println("✓ Dynamic token successfully refreshed!")
	} else if creds.Key == "rds_postgres" {
		fmt.Println("ℹ Using static credentials (no refresh needed)")
	}
}

func getSourceType(key string) string {
	switch key {
	case "dynamic_rds_postgres":
		return "Dynamic RDS (IAM Auth Token)"
	case "rds_postgres":
		return "Static RDS (Password)"
	default:
		return "Unknown"
	}
}

func buildConnectionString(creds *types.Credential) string {
	host := creds.Metadata["host"]
	port := creds.Metadata["port"]
	username := creds.Metadata["username"]
	database := creds.Metadata["database"]
	authMethod := creds.Metadata["auth_method"]

	if authMethod == "iam_token" {
		// For IAM token authentication, the connection string format might be different
		return fmt.Sprintf("postgres://%s@%s:%s/%s?authmethod=iam&token=%s...",
			username, host, port, database, creds.Value[:20])
	} else {
		// Standard password authentication
		return fmt.Sprintf("postgres://%s:***@%s:%s/%s?sslmode=disable",
			username, host, port, database)
	}
}
