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
	"os"
	"strings"

	"go.companyinfo.dev/conflex/codec"
)

// OSEnvVar is a struct that represents an environment variable loader with a prefix.
type OSEnvVar struct {
	prefix  string
	decoder codec.Decoder
}

// NewOSEnvVar creates a new OSEnvVar instance with the given prefix.
func NewOSEnvVar(prefix string) *OSEnvVar {
	return &OSEnvVar{
		prefix:  prefix,
		decoder: codec.EnvVarCodec{},
	}
}

// Load reads the environment variables with the specified prefix and decodes them into a map[string]any.
func (e *OSEnvVar) Load(_ context.Context) (map[string]any, error) {
	validEnv := make([]string, 0, len(os.Environ()))

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, e.prefix) {
			continue
		}

		validEnv = append(validEnv, strings.TrimPrefix(env, e.prefix))
	}

	data := strings.Join(validEnv, "\n")

	var config map[string]any
	if err := e.decoder.Decode([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("error decoding environment variables: %w", err)
	}

	return config, nil
}
