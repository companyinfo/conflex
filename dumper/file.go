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

// Package dumper provides functionality for dumping configuration data to a target.
package dumper

import (
	"context"
	"fmt"
	"os"

	"go.companyinfo.dev/conflex/codec"
)

// File is a struct that represents a file-based configuration dumper.
type File struct {
	path        string
	encoder     codec.Encoder
	permissions os.FileMode
}

const (
	// DefaultFilePermissions represents the default file permissions for dumped configuration files
	DefaultFilePermissions = 0644
)

// NewFile creates a new File instance with the given path and encoder.
// It uses default file permissions of 0644.
func NewFile(path string, encoder codec.Encoder) *File {
	return &File{
		path:        path,
		encoder:     encoder,
		permissions: DefaultFilePermissions,
	}
}

// NewFileWithPermissions creates a new File instance with the given path, encoder, and file permissions.
func NewFileWithPermissions(path string, encoder codec.Encoder, permissions os.FileMode) *File {
	return &File{
		path:        path,
		encoder:     encoder,
		permissions: permissions,
	}
}

// Dump writes the provided values to the file specified by the File instance.
func (f *File) Dump(_ context.Context, values *map[string]any) error {
	data, err := f.encoder.Encode(values)
	if err != nil {
		return fmt.Errorf("failed to encode values: %w", err)
	}

	if err := os.WriteFile(f.path, data, f.permissions); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
