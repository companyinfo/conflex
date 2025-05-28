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

// Package codec provides functionality for encoding and decoding data.
package codec

import (
	"bytes"
	"fmt"
	"strings"
)

// TypeEnvVar is a constant representing the type of an environment variable codec.
const TypeEnvVar Type = "env_var"

// init registers the EnvVarCodec with the codec package under the TypeEnvVar type.
func init() {
	RegisterDecoder(TypeEnvVar, EnvVarCodec{})
}

// EnvVarCodec is a struct that implements the Codec interface for decoding environment variables.
type EnvVarCodec struct{}

// Decode decodes the provided data bytes into a configuration map.
// The data is expected to be in the format of environment variables, with each line containing a key-value pair separated by an equals sign.
func (EnvVarCodec) Decode(data []byte, v any) error {
	conf := make(map[string]any)

	for _, env := range bytes.Split(data, []byte("\n")) {
		pair := strings.SplitN(string(env), "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := pair[0]
		parts := strings.Split(strings.ToLower(key), "_")

		current := conf
		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			}
			if nextMap, ok := current[part].(map[string]any); ok {
				current = nextMap
			} else {
				current[part] = make(map[string]any)
				current = current[part].(map[string]any)
			}
		}

		current[parts[len(parts)-1]] = pair[1]
	}

	ptr, ok := v.(*map[string]any)
	if !ok {
		return fmt.Errorf("EnvVarCodec.Decode: expected *map[string]any, got %T", v)
	}
	*ptr = conf

	return nil
}
