package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.companyinfo.dev/conflex"
	"go.companyinfo.dev/conflex/codec"
)

func TestMixedYAMLAndEnvironmentVariables(t *testing.T) {
	// Set up test environment variables to override YAML defaults
	os.Setenv("WEBAPP_SERVER_PORT", "9090")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_HOST", "test-db")
	os.Setenv("WEBAPP_AUTH_JWT_SECRET", "test-secret")
	os.Setenv("WEBAPP_FEATURES_DEBUG_MODE", "false")

	defer func() {
		os.Unsetenv("WEBAPP_SERVER_PORT")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_HOST")
		os.Unsetenv("WEBAPP_AUTH_JWT_SECRET")
		os.Unsetenv("WEBAPP_FEATURES_DEBUG_MODE")
	}()

	// Create configuration with both YAML and environment variables
	cfg, err := conflex.New(
		conflex.WithFileSource("config.yaml", codec.TypeYAML),
		conflex.WithOSEnvVarSource("WEBAPP_"),
	)
	require.NoError(t, err)

	// Load configuration
	err = cfg.Load(context.Background())
	require.NoError(t, err)

	// Test that environment variables override YAML defaults
	assert.Equal(t, "localhost", cfg.GetString("server.host"))         // From YAML (not overridden)
	assert.Equal(t, 9090, cfg.GetInt("server.port"))                   // From env (overrides YAML's 3000)
	assert.Equal(t, "test-db", cfg.GetString("database.primary.host")) // From env (overrides YAML's localhost)
	assert.Equal(t, 5432, cfg.GetInt("database.primary.port"))         // From YAML (not overridden)
	assert.Equal(t, "test-secret", cfg.GetString("auth.jwt.secret"))   // From env (overrides YAML's dev secret)
	assert.False(t, cfg.GetBool("features.debug.mode"))                // From env (overrides YAML's true)

	// Test struct binding
	var config WebAppConfig
	cfgWithBinding, err := conflex.New(
		conflex.WithFileSource("config.yaml", codec.TypeYAML),
		conflex.WithOSEnvVarSource("WEBAPP_"),
		conflex.WithBinding(&config),
	)
	require.NoError(t, err)

	err = cfgWithBinding.Load(context.Background())
	require.NoError(t, err)

	// Verify struct binding reflects the mixed configuration
	assert.Equal(t, "localhost", config.Server.Host)         // From YAML
	assert.Equal(t, 9090, config.Server.Port)                // From env
	assert.Equal(t, "test-db", config.Database.Primary.Host) // From env
	assert.Equal(t, 5432, config.Database.Primary.Port)      // From YAML
	assert.Equal(t, "test-secret", config.Auth.JWT.Secret)   // From env
	assert.False(t, config.Features.Debug.Mode)              // From env
}

func TestYAMLOnlyConfiguration(t *testing.T) {
	// Clear any existing environment variables
	os.Unsetenv("WEBAPP_SERVER_PORT")
	os.Unsetenv("WEBAPP_DATABASE_PRIMARY_HOST")
	os.Unsetenv("WEBAPP_AUTH_JWT_SECRET")
	os.Unsetenv("WEBAPP_FEATURES_DEBUG_MODE")

	// Create configuration with only YAML file
	cfg, err := conflex.New(
		conflex.WithFileSource("config.yaml", codec.TypeYAML),
	)
	require.NoError(t, err)

	// Load configuration
	err = cfg.Load(context.Background())
	require.NoError(t, err)

	// Test that YAML defaults are used
	assert.Equal(t, "localhost", cfg.GetString("server.host"))
	assert.Equal(t, 3000, cfg.GetInt("server.port"))                     // YAML default
	assert.Equal(t, "localhost", cfg.GetString("database.primary.host")) // YAML default
	assert.Equal(t, 5432, cfg.GetInt("database.primary.port"))
	assert.Equal(t, "dev-jwt-secret-change-in-production", cfg.GetString("auth.jwt.secret")) // YAML default
	assert.True(t, cfg.GetBool("features.debug.mode"))                                       // YAML default
}

func TestEnvironmentVariablesOnly(t *testing.T) {
	// Set environment variables
	os.Setenv("WEBAPP_SERVER_HOST", "env-host")
	os.Setenv("WEBAPP_SERVER_PORT", "8080")
	os.Setenv("WEBAPP_DATABASE_PRIMARY_HOST", "env-db")
	os.Setenv("WEBAPP_AUTH_JWT_SECRET", "env-secret")

	defer func() {
		os.Unsetenv("WEBAPP_SERVER_HOST")
		os.Unsetenv("WEBAPP_SERVER_PORT")
		os.Unsetenv("WEBAPP_DATABASE_PRIMARY_HOST")
		os.Unsetenv("WEBAPP_AUTH_JWT_SECRET")
	}()

	// Create configuration with only environment variables
	cfg, err := conflex.New(
		conflex.WithOSEnvVarSource("WEBAPP_"),
	)
	require.NoError(t, err)

	// Load configuration
	err = cfg.Load(context.Background())
	require.NoError(t, err)

	// Test that environment variables are used
	assert.Equal(t, "env-host", cfg.GetString("server.host"))
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
	assert.Equal(t, "env-db", cfg.GetString("database.primary.host"))
	assert.Equal(t, "env-secret", cfg.GetString("auth.jwt.secret"))
}
