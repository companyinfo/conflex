// Copyright 2025 Company.info B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package source provides functionality for loading configuration data from various sources.
package source

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"go.companyinfo.dev/conflex/codec"
)

// ConsulKV is an interface for Consul KV operations (for testability)
type ConsulKV interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
}

// Consul is a struct that represents a Consul-based configuration source.
type Consul struct {
	client    *api.Client
	kv        ConsulKV
	path      string
	lastIndex uint64
	decoder   codec.Decoder
}

// NewConsul creates a new Consul configuration source with the given path and decoder.
// If kv is nil, it uses the default client.KV().
func NewConsul(path string, decoder codec.Decoder, kv ConsulKV) (*Consul, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}
	if kv == nil {
		kv = client.KV()
	}
	return &Consul{
		client:  client,
		kv:      kv,
		path:    path,
		decoder: decoder,
	}, nil
}

// Load retrieves the configuration data from the Consul key-value store at the specified path.
func (c *Consul) Load(ctx context.Context) (map[string]any, error) {
	pair, meta, err := c.kv.Get(c.path, (&api.QueryOptions{}).WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get consul key: %w", err)
	}

	if pair == nil {
		return make(map[string]any), nil
	}

	// Only update lastIndex if meta is not nil
	if meta != nil {
		c.lastIndex = meta.LastIndex
	}

	var config map[string]any
	caster, ok := c.decoder.(*codec.CasterCodec)
	if ok {
		var val any

		// If the decoder is a caster, the config key is the last part of the Consul key
		keyParts := strings.Split(pair.Key, "/")
		key := keyParts[len(keyParts)-1]

		err = caster.Decode(pair.Value, &val)
		if err != nil {
			return nil, fmt.Errorf("failed to decode consul value: %w", err)
		}

		config = map[string]any{key: val}
		return config, nil
	}

	if err := c.decoder.Decode(pair.Value, &config); err != nil {
		return nil, fmt.Errorf("failed to decode consul value: %w", err)
	}

	return config, nil
}
