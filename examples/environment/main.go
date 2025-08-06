package main

import (
	"context"
	"fmt"
	"log"

	"go.companyinfo.dev/conflex"
)

// SimpleConfig represents a simple configuration without validation
type SimpleConfig struct {
	Server   ServerConfig   `conflex:"server"`
	Database DatabaseConfig `conflex:"database"`
	Auth     AuthConfig     `conflex:"auth"`
	Features FeaturesConfig `conflex:"features"`
}

type ServerConfig struct {
	Host string `conflex:"host"`
	Port int    `conflex:"port"`
}

type DatabaseConfig struct {
	Primary PrimaryConfig `conflex:"primary"`
}

type PrimaryConfig struct {
	Host     string `conflex:"host"`
	Port     int    `conflex:"port"`
	Database string `conflex:"database"`
}

type AuthConfig struct {
	JWT JWTConfig `conflex:"jwt"`
}

type JWTConfig struct {
	Secret string `conflex:"secret"`
}

type FeaturesConfig struct {
	Debug DebugConfig `conflex:"debug"`
}

type DebugConfig struct {
	Mode bool `conflex:"mode"`
}

// PrintConfig displays the configuration in a readable format
func (c *SimpleConfig) PrintConfig() {
	fmt.Println("=== Simple Configuration ===")
	fmt.Printf("Server: %s:%d\n", c.Server.Host, c.Server.Port)
	fmt.Printf("Database: %s:%d/%s\n", c.Database.Primary.Host, c.Database.Primary.Port, c.Database.Primary.Database)
	fmt.Printf("Auth JWT Secret: %s\n", c.Auth.JWT.Secret)
	fmt.Printf("Debug Mode: %t\n", c.Features.Debug.Mode)
	fmt.Println("============================")
}

func main() {
	var config SimpleConfig

	// Create configuration with environment variable source
	cfg, err := conflex.New(
		conflex.WithOSEnvVarSource("WEBAPP_"),
		conflex.WithBinding(&config),
	)
	if err != nil {
		log.Fatalf("Failed to create configuration: %v", err)
	}

	// Load configuration
	if err := cfg.Load(context.Background()); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print the loaded configuration
	config.PrintConfig()

	// Demonstrate accessing configuration values directly
	fmt.Println("\n=== Direct Configuration Access ===")
	serverHost := cfg.GetString("server.host")
	serverPort := cfg.GetInt("server.port")
	databaseHost := cfg.GetString("database.primary.host")

	fmt.Printf("Server: %s:%d\n", serverHost, serverPort)
	fmt.Printf("Database: %s\n", databaseHost)

	// Check if debug mode is enabled
	if debugMode := cfg.GetBool("features.debug.mode"); debugMode {
		fmt.Println("Debug mode is enabled")
	} else {
		fmt.Println("Debug mode is disabled")
	}
}
