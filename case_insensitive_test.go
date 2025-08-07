package conflex

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.companyinfo.dev/conflex/codec"
)

func TestCaseInsensitiveMerging(t *testing.T) {
	// Test data with mixed case keys
	config1 := []byte(`{
		"Server": {
			"Host": "localhost",
			"Port": 8080
		},
		"Database": {
			"Name": "testdb"
		}
	}`)

	config2 := []byte(`{
		"server": {
			"host": "example.com",
			"port": 9090
		},
		"database": {
			"name": "prod"
		}
	}`)

	// Create configuration with both sources
	cfg, err := New(
		WithContentSource(config1, codec.TypeJSON),
		WithContentSource(config2, codec.TypeJSON),
	)
	require.NoError(t, err)

	// Load configuration
	err = cfg.Load(context.Background())
	require.NoError(t, err)

	// Test case-insensitive access - should all work regardless of case
	assert.Equal(t, "example.com", cfg.GetString("server.host"))
	assert.Equal(t, "example.com", cfg.GetString("Server.Host"))
	assert.Equal(t, "example.com", cfg.GetString("SERVER.HOST"))

	assert.Equal(t, 9090, cfg.GetInt("server.port"))
	assert.Equal(t, 9090, cfg.GetInt("Server.Port"))
	assert.Equal(t, 9090, cfg.GetInt("SERVER.PORT"))

	assert.Equal(t, "prod", cfg.GetString("database.name"))
	assert.Equal(t, "prod", cfg.GetString("Database.Name"))
	assert.Equal(t, "prod", cfg.GetString("DATABASE.NAME"))
}

func TestNormalizeMapKeys(t *testing.T) {
	input := map[string]any{
		"Server": map[string]any{
			"Host": "localhost",
			"Port": 8080,
		},
		"Database": map[string]any{
			"Name": "testdb",
			"Settings": map[string]any{
				"MaxConnections": 100,
			},
		},
	}

	normalized := normalizeMapKeys(input)

	expected := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
		"database": map[string]any{
			"name": "testdb",
			"settings": map[string]any{
				"maxconnections": 100,
			},
		},
	}

	assert.Equal(t, expected, normalized)
}
