package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.companyinfo.dev/conflex"
	"go.companyinfo.dev/conflex/codec"
)

// WebAppConfig represents a web application configuration without validation
type WebAppConfig struct {
	Server     ServerConfig     `conflex:"server"`
	Database   DatabaseConfig   `conflex:"database"`
	Redis      RedisConfig      `conflex:"redis"`
	Auth       AuthConfig       `conflex:"auth"`
	Logging    LoggingConfig    `conflex:"logging"`
	Monitoring MonitoringConfig `conflex:"monitoring"`
	Features   FeaturesConfig   `conflex:"features"`
}

type ServerConfig struct {
	Host         string        `conflex:"host"`
	Port         int           `conflex:"port"`
	ReadTimeout  time.Duration `conflex:"read.timeout"`
	WriteTimeout time.Duration `conflex:"write.timeout"`
	TLS          TLSConfig     `conflex:"tls"`
}

type TLSConfig struct {
	Enabled bool       `conflex:"enabled"`
	Cert    CertConfig `conflex:"cert"`
	Key     KeyConfig  `conflex:"key"`
}

type CertConfig struct {
	File string `conflex:"file"`
}

type KeyConfig struct {
	File string `conflex:"file"`
}

type DatabaseConfig struct {
	Primary PrimaryConfig `conflex:"primary"`
	Replica ReplicaConfig `conflex:"replica"`
	Pool    PoolConfig    `conflex:"pool"`
}

type PrimaryConfig struct {
	Host     string `conflex:"host"`
	Port     int    `conflex:"port"`
	Database string `conflex:"database"`
	Username string `conflex:"username"`
	Password string `conflex:"password"`
	SSLMode  string `conflex:"ssl.mode"`
}

type ReplicaConfig struct {
	Host     string `conflex:"host"`
	Port     int    `conflex:"port"`
	Database string `conflex:"database"`
	Username string `conflex:"username"`
	Password string `conflex:"password"`
	SSLMode  string `conflex:"ssl.mode"`
}

type PoolConfig struct {
	Max MaxConfig `conflex:"max"`
}

type MaxConfig struct {
	Open     int           `conflex:"open"`
	Idle     int           `conflex:"idle"`
	Lifetime time.Duration `conflex:"lifetime"`
}

type RedisConfig struct {
	Host     string        `conflex:"host"`
	Port     int           `conflex:"port"`
	Password string        `conflex:"password"`
	Database int           `conflex:"database"`
	Timeout  time.Duration `conflex:"timeout"`
}

type AuthConfig struct {
	JWT     JWTConfig     `conflex:"jwt"`
	Token   TokenConfig   `conflex:"token"`
	Refresh RefreshConfig `conflex:"refresh"`
}

type JWTConfig struct {
	Secret string `conflex:"secret"`
}

type TokenConfig struct {
	Duration time.Duration `conflex:"duration"`
}

type RefreshConfig struct {
	Secret string `conflex:"secret"`
}

type LoggingConfig struct {
	Level      string `conflex:"level"`
	Format     string `conflex:"format"`
	OutputFile string `conflex:"output.file"`
}

type MonitoringConfig struct {
	Enabled     bool   `conflex:"enabled"`
	MetricsPort int    `conflex:"metrics.port"`
	HealthPath  string `conflex:"health.path"`
}

type FeaturesConfig struct {
	RateLimit RateLimitConfig `conflex:"rate.limit"`
	Cache     CacheConfig     `conflex:"cache"`
	Debug     DebugConfig     `conflex:"debug"`
}

type RateLimitConfig struct {
	Enabled bool `conflex:"enabled"`
}

type CacheConfig struct {
	Enabled bool `conflex:"enabled"`
}

type DebugConfig struct {
	Mode bool `conflex:"mode"`
}

// PrintConfig displays the configuration in a readable format
func (c *WebAppConfig) PrintConfig() {
	fmt.Println("=== Web Application Configuration (YAML + Environment Variables) ===")
	fmt.Printf("Server: %s:%d\n", c.Server.Host, c.Server.Port)
	fmt.Printf("  Read Timeout: %v\n", c.Server.ReadTimeout)
	fmt.Printf("  Write Timeout: %v\n", c.Server.WriteTimeout)
	fmt.Printf("  TLS Enabled: %t\n", c.Server.TLS.Enabled)
	if c.Server.TLS.Enabled {
		fmt.Printf("  TLS Cert: %s\n", c.Server.TLS.Cert.File)
		fmt.Printf("  TLS Key: %s\n", c.Server.TLS.Key.File)
	}

	fmt.Printf("\nDatabase Primary: %s:%d/%s\n",
		c.Database.Primary.Host, c.Database.Primary.Port, c.Database.Primary.Database)
	fmt.Printf("Database Replica: %s:%d/%s\n",
		c.Database.Replica.Host, c.Database.Replica.Port, c.Database.Replica.Database)
	fmt.Printf("Database Pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v\n",
		c.Database.Pool.Max.Open, c.Database.Pool.Max.Idle, c.Database.Pool.Max.Lifetime)

	fmt.Printf("\nRedis: %s:%d (DB: %d)\n", c.Redis.Host, c.Redis.Port, c.Redis.Database)
	fmt.Printf("Redis Timeout: %v\n", c.Redis.Timeout)

	fmt.Printf("\nAuth Token Duration: %v\n", c.Auth.Token.Duration)
	fmt.Printf("Logging Level: %s, Format: %s\n", c.Logging.Level, c.Logging.Format)
	if c.Logging.OutputFile != "" {
		fmt.Printf("Logging Output: %s\n", c.Logging.OutputFile)
	}

	fmt.Printf("\nMonitoring Enabled: %t\n", c.Monitoring.Enabled)
	if c.Monitoring.Enabled {
		fmt.Printf("Metrics Port: %d\n", c.Monitoring.MetricsPort)
		fmt.Printf("Health Path: %s\n", c.Monitoring.HealthPath)
	}

	fmt.Printf("\nFeatures:\n")
	fmt.Printf("  Rate Limit: %t\n", c.Features.RateLimit.Enabled)
	fmt.Printf("  Cache: %t\n", c.Features.Cache.Enabled)
	fmt.Printf("  Debug Mode: %t\n", c.Features.Debug.Mode)
	fmt.Println("=====================================")
}

func main() {
	var config WebAppConfig

	// Create configuration with multiple sources
	cfg, err := conflex.New(
		// First, load from YAML file (default values)
		conflex.WithFileSource("config.yaml", codec.TypeYAML),
		// Then, override with environment variables (higher precedence)
		conflex.WithOSEnvVarSource("WEBAPP_"),
		// Bind to our struct
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

	// Check if TLS is enabled
	if tlsEnabled := cfg.GetBool("server.tls.enabled"); tlsEnabled {
		fmt.Println("TLS is enabled")
	} else {
		fmt.Println("TLS is disabled")
	}

	// Demonstrate configuration precedence
	fmt.Println("\n=== Configuration Precedence Demo ===")
	fmt.Println("Values are loaded in this order:")
	fmt.Println("1. YAML file (config.yaml) - default values")
	fmt.Println("2. Environment variables (WEBAPP_*) - override defaults")
	fmt.Println("")
	fmt.Println("Example: If YAML has server.port=3000 and env has WEBAPP_SERVER_PORT=8080")
	fmt.Println("The final value will be 8080 (environment variable wins)")
}
