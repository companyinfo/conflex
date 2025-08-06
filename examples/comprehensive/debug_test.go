package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.companyinfo.dev/conflex"
	"go.companyinfo.dev/conflex/source"
)

func TestDebugEnvVars(t *testing.T) {
	// Set a few environment variables
	os.Setenv("WEBAPP_SERVER_HOST", "test-host")
	os.Setenv("WEBAPP_SERVER_PORT", "8080")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_HOST", "test-db")

	defer func() {
		os.Unsetenv("WEBAPP_SERVER_HOST")
		os.Unsetenv("WEBAPP_SERVER_PORT")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_HOST")
	}()

	// Create environment variable source directly
	envSource := source.NewOSEnvVar("WEBAPP_")

	// Load configuration
	config, err := envSource.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load environment variables: %v", err)
	}

	fmt.Printf("Loaded config: %+v\n", config)

	// Check specific values
	if host, ok := config["server"].(map[string]any); ok {
		if serverHost, exists := host["host"]; exists {
			fmt.Printf("Server host: %v\n", serverHost)
		} else {
			fmt.Printf("Server host not found in: %+v\n", host)
		}
	} else {
		fmt.Printf("Server section not found or not a map: %T %+v\n", config["server"], config["server"])
	}

	// Test with conflex
	cfg, err := conflex.New(
		conflex.WithOSEnvVarSource("WEBAPP_"),
	)
	if err != nil {
		t.Fatalf("Failed to create conflex: %v", err)
	}

	err = cfg.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load conflex: %v", err)
	}

	fmt.Printf("Conflex server host: %s\n", cfg.GetString("server.host"))
	fmt.Printf("Conflex server port: %d\n", cfg.GetInt("server.port"))
	fmt.Printf("Conflex database host: %s\n", cfg.GetString("database.primary.host"))
}
