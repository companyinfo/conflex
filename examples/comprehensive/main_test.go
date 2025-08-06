package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.companyinfo.dev/conflex"
)

func TestWebAppConfig_EnvironmentVariables(t *testing.T) {
	// Set up test environment variables (all required fields)
	os.Setenv("WEBAPP_SERVER_HOST", "test-host")
	os.Setenv("WEBAPP_SERVER_PORT", "9090")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_HOST", "test-db")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_PORT", "5432")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_DATABASE", "testdb")
	os.Setenv("WEBAPP_AUTH_JWT_SECRET", "test-secret")
	os.Setenv("WEBAPP_AUTH_TOKEN_DURATION", "1h")
	os.Setenv("WEBAPP_FEATURES_DEBUG_MODE", "true")

	// Debug: Check if environment variables are set
	fmt.Printf("WEBAPP_SERVER_HOST: %s\n", os.Getenv("WEBAPP_SERVER_HOST"))
	fmt.Printf("WEBAPP_AUTH_JWT_SECRET: %s\n", os.Getenv("WEBAPP_AUTH_JWT_SECRET"))

	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("WEBAPP_SERVER_HOST")
		os.Unsetenv("WEBAPP_SERVER_PORT")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_HOST")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_PORT")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_DATABASE")
		os.Unsetenv("WEBAPP_AUTH_JWT_SECRET")
		os.Unsetenv("WEBAPP_AUTH_TOKEN_DURATION")
		os.Unsetenv("WEBAPP_FEATURES_DEBUG_MODE")
	}()

	// Create configuration without binding to test direct access
	cfg, err := conflex.New(
		conflex.WithOSEnvVarSource("WEBAPP_"),
	)
	require.NoError(t, err)

	// Load configuration
	err = cfg.Load(context.Background())
	require.NoError(t, err)

	// Debug: Check what was loaded
	fmt.Printf("Loaded server.host: %s\n", cfg.GetString("server.host"))
	fmt.Printf("Loaded auth.jwt.secret: %s\n", cfg.GetString("auth.jwt.secret"))
	fmt.Printf("Loaded features.debug.mode: %t\n", cfg.GetBool("features.debug.mode"))

	// Test direct configuration access
	assert.Equal(t, "test-host", cfg.GetString("server.host"))
	assert.Equal(t, 9090, cfg.GetInt("server.port"))
	assert.Equal(t, "test-db", cfg.GetString("database.primary.host"))
	assert.Equal(t, 5432, cfg.GetInt("database.primary.port"))
	assert.Equal(t, "testdb", cfg.GetString("database.primary.database"))
	assert.Equal(t, "test-secret", cfg.GetString("auth.jwt.secret"))
	assert.True(t, cfg.GetBool("features.debug.mode"))

	// Now test with binding
	var config WebAppConfig
	cfgWithBinding, err := conflex.New(
		conflex.WithOSEnvVarSource("WEBAPP_"),
		conflex.WithBinding(&config),
	)
	require.NoError(t, err)

	// Load configuration with binding
	err = cfgWithBinding.Load(context.Background())
	require.NoError(t, err)

	// Test struct binding
	assert.Equal(t, "test-host", config.Server.Host)
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "test-db", config.Database.Primary.Host)
	assert.Equal(t, 5432, config.Database.Primary.Port)
	assert.Equal(t, "testdb", config.Database.Primary.Database)
	assert.Equal(t, "test-secret", config.Auth.JWT.Secret)
	assert.True(t, config.Features.Debug.Mode)
}

func TestWebAppConfig_NestedStructures(t *testing.T) {
	// Test nested environment variable mapping (including required fields)
	os.Setenv("WEBAPP_SERVER_HOST", "test-host")
	os.Setenv("WEBAPP_SERVER_PORT", "9090")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_HOST", "test-db")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_PORT", "5432")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_DATABASE", "testdb")
	os.Setenv("WEBAPP_AUTH_JWT_SECRET", "test-secret")
	os.Setenv("WEBAPP_AUTH_TOKEN_DURATION", "1h")

	os.Setenv("WEBAPP_SERVER_TLS_ENABLED", "true")
	os.Setenv("WEBAPP_SERVER_TLS_CERT_FILE", "/path/to/cert.pem")
	os.Setenv("WEBAPP_SERVER_TLS_KEY_FILE", "/path/to/key.pem")
	os.Setenv("WEBAPP_DATABASE_POOL_MAX_OPEN", "50")
	os.Setenv("WEBAPP_DATABASE_POOL_MAX_IDLE", "10")

	defer func() {
		os.Unsetenv("WEBAPP_SERVER_HOST")
		os.Unsetenv("WEBAPP_SERVER_PORT")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_HOST")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_PORT")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_DATABASE")
		os.Unsetenv("WEBAPP_AUTH_JWT_SECRET")
		os.Unsetenv("WEBAPP_AUTH_TOKEN_DURATION")
		os.Unsetenv("WEBAPP_SERVER_TLS_ENABLED")
		os.Unsetenv("WEBAPP_SERVER_TLS_CERT_FILE")
		os.Unsetenv("WEBAPP_SERVER_TLS_KEY_FILE")
		os.Unsetenv("WEBAPP_DATABASE_POOL_MAX_OPEN")
		os.Unsetenv("WEBAPP_DATABASE_POOL_MAX_IDLE")
	}()

	// Test direct access first
	cfg, err := conflex.New(
		conflex.WithOSEnvVarSource("WEBAPP_"),
	)
	require.NoError(t, err)

	err = cfg.Load(context.Background())
	require.NoError(t, err)

	// Test direct access to nested values
	assert.True(t, cfg.GetBool("server.tls.enabled"))
	assert.Equal(t, "/path/to/cert.pem", cfg.GetString("server.tls.cert.file"))
	assert.Equal(t, "/path/to/key.pem", cfg.GetString("server.tls.key.file"))
	assert.Equal(t, 50, cfg.GetInt("database.pool.max.open"))
	assert.Equal(t, 10, cfg.GetInt("database.pool.max.idle"))

	// Now test with binding
	var config WebAppConfig
	cfgWithBinding, err := conflex.New(
		conflex.WithOSEnvVarSource("WEBAPP_"),
		conflex.WithBinding(&config),
	)
	require.NoError(t, err)

	err = cfgWithBinding.Load(context.Background())
	require.NoError(t, err)

	// Test nested TLS configuration
	assert.True(t, config.Server.TLS.Enabled)
	assert.Equal(t, "/path/to/cert.pem", config.Server.TLS.Cert.File)
	assert.Equal(t, "/path/to/key.pem", config.Server.TLS.Key.File)

	// Test nested database pool configuration
	assert.Equal(t, 50, config.Database.Pool.Max.Open)
	assert.Equal(t, 10, config.Database.Pool.Max.Idle)
}
