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
	"github.com/fsnotify/fsnotify"
	"go.companyinfo.dev/conflex/codec"
	"log"
	"os"
	"time"
)

// File represents a configuration file that can be loaded and watched for changes.
type File struct {
	path    string
	data    []byte
	decoder codec.Decoder
}

// NewFile creates a new File instance with the given path and decoder.
func NewFile(path string, decoder codec.Decoder) *File {
	return &File{
		path:    path,
		decoder: decoder,
	}
}

// NewFileContent creates a new File instance with the given data and decoder.
func NewFileContent(data []byte, decoder codec.Decoder) *File {
	return &File{
		data:    data,
		decoder: decoder,
	}
}

// Load reads the configuration file and decodes its contents into a map[string]any.
func (f *File) Load(context.Context) (map[string]any, error) {
	var err error

	if f.path != "" {
		f.data, err = os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
	}

	var config map[string]any
	if err := f.decoder.Decode(f.data, &config); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}

	return config, nil
}

// Watch monitors the configuration file for changes and reloads the file when it is modified.
func (f *File) Watch(_ context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			fmt.Println("Failed to close watcher:", err)
		}
	}(watcher)

	if err := watcher.Add(f.path); err != nil {
		return err
	}

	var timer *time.Timer
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(100*time.Millisecond, func() {
					log.Println("Reloading file:", f.path)
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher closed")
			}
			log.Println("error:", err)
		}
	}
}
