package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jayxuchen/credmanager"
	"github.com/jayxuchen/credmanager/sources"
	"github.com/jayxuchen/credmanager/sources/static"
)

func main() {
	// Create a static RDS Postgres credential source
	rdsSource := static.NewRDSPostgresSource(
		"localhost",     // host
		5432,           // port
		"myuser",       // username
		"mypassword",   // password
		"mydatabase",   // database
	)

	// Create credential manager with the RDS source
	sources := []sources.CredentialSource{rdsSource}
	manager := credmanager.NewManager(sources)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get credentials from the first valid source
	creds, err := manager.GetFirstValid(ctx)
	if err != nil {
		log.Fatalf("Failed to get credentials: %v", err)
	}

	// Print credential information
	fmt.Printf("Credential Key: %s\n", creds.Key)
	fmt.Printf("Credential Value: %s\n", creds.Value)
	
	if creds.Expiry != nil {
		fmt.Printf("Expires at: %s\n", creds.Expiry.Format(time.RFC3339))
	} else {
		fmt.Println("Expiry: Never")
	}

	fmt.Println("\nMetadata:")
	for key, value := range creds.Metadata {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// Example of building a connection string from the metadata
	host := creds.Metadata["host"]
	port := creds.Metadata["port"]
	username := creds.Metadata["username"]
	database := creds.Metadata["database"]
	password := creds.Value

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, host, port, database)
	
	fmt.Printf("\nConnection String: %s\n", connectionString)
}
