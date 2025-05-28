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

// Package conflex provides a flexible configuration package.
package conflex

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cast"
	"go.companyinfo.dev/conflex/codec"
	"go.companyinfo.dev/conflex/dumper"
	"go.companyinfo.dev/conflex/source"
)

// Option is a functional option that can be used to configure a Conflex instance.
type Option func(c *Conflex) error

// Conflex is the main struct that holds the configuration data and the sources used to load the data.
// The values field is a pointer to a map that holds the configuration data.
// The sources field is a slice of Source instances that are used to load the configuration data.
// The mu field is a sync.RWMutex that is used to synchronize access to the configuration data.
type Conflex struct {
	values  *map[string]any
	sources []Source
	dumpers []Dumper
	binding any
	mu      sync.RWMutex
}

// WithSource returns an Option that configures the Conflex instance to add a source for loading configuration data.
func WithSource(loader Source) Option {
	return func(c *Conflex) error {
		c.sources = append(c.sources, loader)
		return nil
	}
}

// WithDumper returns an Option that configures the Conflex instance to add a dumper for configuration data.
func WithDumper(dumper Dumper) Option {
	return func(c *Conflex) error {
		c.dumpers = append(c.dumpers, dumper)
		return nil
	}
}

// WithFileDumper returns an Option that configures the Conflex instance to dump configuration data to a file.
func WithFileDumper(path string, codecType codec.Type) Option {
	return func(c *Conflex) error {
		encoder, err := codec.GetEncoder(codecType)
		if err != nil {
			return fmt.Errorf("failed to get encoder: %w", err)
		}

		c.dumpers = append(c.dumpers, dumper.NewFile(path, encoder))
		return nil
	}
}

// WithFileSource returns an Option that configures the Conflex instance to load configuration data from a file.
func WithFileSource(path string, codecType codec.Type) Option {
	return func(c *Conflex) error {
		decoder, err := codec.GetDecoder(codecType)
		if err != nil {
			return fmt.Errorf("failed to get decoder: %w", err)
		}

		c.sources = append(c.sources, source.NewFile(path, decoder))
		return nil
	}
}

// WithContentSource returns an Option that configures the Conflex instance to load configuration data from a byte slice.
func WithContentSource(data []byte, codecType codec.Type) Option {
	return func(c *Conflex) error {
		decoder, err := codec.GetDecoder(codecType)
		if err != nil {
			return fmt.Errorf("failed to get decoder: %w", err)
		}

		c.sources = append(c.sources, source.NewFileContent(data, decoder))
		return nil
	}
}

// WithOSEnvVarSource returns an Option that configures the Conflex instance to load configuration data from environment variables.
// The prefix parameter specifies the prefix for the environment variables to be loaded.
func WithOSEnvVarSource(prefix string) Option {
	return func(c *Conflex) error {
		c.sources = append(c.sources, source.NewOSEnvVar(prefix))
		return nil
	}
}

// WithConsulSource returns an Option that configures the Conflex instance to load configuration data from a Consul server.
// The path parameter specifies the key path in Consul's key-value store to load configuration from.
// The codecType parameter specifies the codec type (e.g., JSON, YAML) to use for decoding the configuration data.
// Required environment variables:
//   - CONSUL_HTTP_ADDR: The address of the Consul server (e.g., "http://localhost:8500")
//   - CONSUL_HTTP_TOKEN: The access token for authentication with Consul (optional)
func WithConsulSource(path string, codecType codec.Type) Option {
	return func(c *Conflex) error {
		decoder, err := codec.GetDecoder(codecType)
		if err != nil {
			return fmt.Errorf("failed to get decoder: %w", err)
		}

		l, err := source.NewConsul(path, decoder, nil)
		if err != nil {
			return err
		}

		c.sources = append(c.sources, l)

		return nil
	}
}

// WithBinding returns an Option that configures the Conflex instance to bind the configuration data to a struct.
func WithBinding(v any) Option {
	return func(c *Conflex) error {
		c.binding = v

		return nil
	}
}

// New creates a new Conflex instance with the provided options.
// It iterates through the options and applies each one to the Conflex instance.
// If any of the options return an error, the errors are collected and returned.
func New(options ...Option) (*Conflex, error) {
	var errs error
	c := &Conflex{
		values:  &map[string]any{},
		sources: []Source{},
	}

	for _, option := range options {
		err := option(c)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return c, errs
}

// Load loads configuration data from the registered sources and merges it into the internal values map.
// The method acquires a write lock on the values map before loading the configuration data, and releases the lock before returning.
// If any of the sources fail to load the configuration data, the method returns the first encountered error.
func (c *Conflex) Load(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, l := range c.sources {
		conf, err := l.Load(ctx)
		if err != nil {
			return err
		}

		err = mergo.Merge(c.values, conf, mergo.WithOverride)
		if err != nil {
			return err
		}
	}

	if c.binding != nil {
		if err := c.bind(); err != nil {
			return err
		}
	}

	return nil
}

// Dump writes the current configuration values to the registered dumpers.
func (c *Conflex) Dump(ctx context.Context) error {
	for _, d := range c.dumpers {
		if err := d.Dump(ctx, c.Values()); err != nil {
			return err
		}
	}

	return nil
}

func (c *Conflex) bind() error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "config", Result: c.binding})
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(c.values); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	return err
}

// Values returns a pointer to the internal values map of the Conflex instance.
// The map is protected by a read lock, which is acquired and released within this method.
// This method is used to safely access the internal values map.
func (c *Conflex) Values() *map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.values
}

// getValueFromMap retrieves the value associated with the given path from the internal values map.
// The path is a dot-separated string that represents the nested structure of the map.
// If the path is valid and the final value is found, it is returned. Otherwise, nil is returned.
func (c *Conflex) getValueFromMap(path string) any {
	segments := strings.Split(path, ".")
	current := c.Values()

	for i, segment := range segments {
		if currentMap, ok := (*current)[segment]; ok {
			// If it's the last segment, return the value
			if i == len(segments)-1 {
				return currentMap
			}

			// If it's a nested map, continue traversal
			if nestedMap, ok := currentMap.(map[string]any); ok {
				current = &nestedMap
			} else {
				// The path is invalid if a segment is not a map
				return nil
			}
		} else {
			// Key does not exist
			return nil
		}
	}

	return nil
}

// Get returns the value associated with the given key as an any type.
// If the key is not found, it returns nil.
func (c *Conflex) Get(key string) any {
	return c.getValueFromMap(key)
}

// GetString returns the value associated with the given key as a string.
// If the value is not found or cannot be converted to a string, an empty string is returned.
func (c *Conflex) GetString(key string) string {
	return cast.ToString(c.Get(key))
}

// GetStringE returns the value associated with the given key as a string.
// If the value is not found or cannot be converted to a string, it returns an error.
func (c *Conflex) GetStringE(key string) (string, error) {
	return cast.ToStringE(c.Get(key))
}

// GetBool returns the value associated with the given key as a boolean.
// If the value is not found or cannot be converted to a boolean, false is returned.
func (c *Conflex) GetBool(key string) bool {
	return cast.ToBool(c.Get(key))
}

// GetBoolE returns the value associated with the given key as a boolean.
// If the value is not found or cannot be converted to a boolean, it returns an error.
func (c *Conflex) GetBoolE(key string) (bool, error) {
	return cast.ToBoolE(c.Get(key))
}

// GetInt returns the value associated with the given key as an integer.
// If the value is not found or cannot be converted to an integer, 0 is returned.
func (c *Conflex) GetInt(key string) int {
	return cast.ToInt(c.Get(key))
}

// GetIntE returns the value associated with the given key as an integer.
// If the value is not found or cannot be converted to an integer, it returns an error.
func (c *Conflex) GetIntE(key string) (int, error) {
	return cast.ToIntE(c.Get(key))
}

// GetInt32 returns the value associated with the given key as an int32.
// If the value is not found or cannot be converted to an int32, 0 is returned.
func (c *Conflex) GetInt32(key string) int32 {
	return cast.ToInt32(c.Get(key))
}

// GetInt32E returns the value associated with the given key as an int32.
// If the value is not found or cannot be converted to an int32, it returns an error.
func (c *Conflex) GetInt32E(key string) (int32, error) {
	return cast.ToInt32E(c.Get(key))
}

// GetInt64 returns the value associated with the given key as an int64.
// If the value is not found or cannot be converted to an int64, 0 is returned.
func (c *Conflex) GetInt64(key string) int64 {
	return cast.ToInt64(c.Get(key))
}

// GetInt64E returns the value associated with the given key as an int64.
// If the value is not found or cannot be converted to an int64, it returns an error.
func (c *Conflex) GetInt64E(key string) (int64, error) {
	return cast.ToInt64E(c.Get(key))
}

// GetUint8 returns the value associated with the given key as an uint8.
// If the value is not found or cannot be converted to an uint8, 0 is returned.
func (c *Conflex) GetUint8(key string) uint8 {
	return cast.ToUint8(c.Get(key))
}

// GetUint8E returns the value associated with the given key as an uint8.
// If the value is not found or cannot be converted to an uint8, it returns an error.
func (c *Conflex) GetUint8E(key string) (uint8, error) {
	return cast.ToUint8E(c.Get(key))
}

// GetUint returns the value associated with the given key as an uint.
// If the value is not found or cannot be converted to an uint, 0 is returned.
func (c *Conflex) GetUint(key string) uint {
	return cast.ToUint(c.Get(key))
}

// GetUintE returns the value associated with the given key as an uint.
// If the value is not found or cannot be converted to an uint, it returns an error.
func (c *Conflex) GetUintE(key string) (uint, error) {
	return cast.ToUintE(c.Get(key))
}

// GetUint16 returns the value associated with the given key as an uint16.
// If the value is not found or cannot be converted to an uint16, 0 is returned.
func (c *Conflex) GetUint16(key string) uint16 {
	return cast.ToUint16(c.Get(key))
}

// GetUint16E returns the value associated with the given key as an uint16.
// If the value is not found or cannot be converted to an uint16, it returns an error.
func (c *Conflex) GetUint16E(key string) (uint16, error) {
	return cast.ToUint16E(c.Get(key))
}

// GetUint32 returns the value associated with the given key as an uint32.
// If the value is not found or cannot be converted to an uint32, 0 is returned.
func (c *Conflex) GetUint32(key string) uint32 {
	return cast.ToUint32(c.Get(key))
}

// GetUint32E returns the value associated with the given key as an uint32.
// If the value is not found or cannot be converted to an uint32, it returns an error.
func (c *Conflex) GetUint32E(key string) (uint32, error) {
	return cast.ToUint32E(c.Get(key))
}

// GetUint64 returns the value associated with the given key as an uint64.
// If the value is not found or cannot be converted to an uint64, 0 is returned.
func (c *Conflex) GetUint64(key string) uint64 {
	return cast.ToUint64(c.Get(key))
}

// GetUint64E returns the value associated with the given key as an uint64.
// If the value is not found or cannot be converted to an uint64, it returns an error.
func (c *Conflex) GetUint64E(key string) (uint64, error) {
	return cast.ToUint64E(c.Get(key))
}

// GetFloat64 returns the value associated with the given key as a float64.
// If the value is not found or cannot be converted to a float64, 0.0 is returned.
func (c *Conflex) GetFloat64(key string) float64 {
	return cast.ToFloat64(c.Get(key))
}

// GetFloat64E returns the value associated with the given key as a float64.
// If the value is not found or cannot be converted to a float64, it returns an error.
func (c *Conflex) GetFloat64E(key string) (float64, error) {
	return cast.ToFloat64E(c.Get(key))
}

// GetTime returns the value associated with the given key as a time.Time.
// If the value is not found or cannot be converted to a time.Time, the zero value is returned.
func (c *Conflex) GetTime(key string) time.Time {
	return cast.ToTime(c.Get(key))
}

// GetTimeE returns the value associated with the given key as a time.Time.
// If the value is not found or cannot be converted to a time.Time, it returns an error.
func (c *Conflex) GetTimeE(key string) (time.Time, error) {
	return cast.ToTimeE(c.Get(key))
}

// GetDuration returns the value associated with the given key as a time.Duration.
// If the value is not found or cannot be converted to a time.Duration, the zero value is returned.
func (c *Conflex) GetDuration(key string) time.Duration {
	return cast.ToDuration(c.Get(key))
}

// GetDurationE returns the value associated with the given key as a time.Duration.
// If the value is not found or cannot be converted to a time.Duration, it returns an error.
func (c *Conflex) GetDurationE(key string) (time.Duration, error) {
	return cast.ToDurationE(c.Get(key))
}

// GetIntSlice returns the value associated with the given key as a slice of integers.
// If the value is not found or cannot be converted to a slice of integers, an empty slice is returned.
func (c *Conflex) GetIntSlice(key string) []int {
	return cast.ToIntSlice(c.Get(key))
}

// GetIntSliceE returns the value associated with the given key as a slice of integers.
// If the value is not found or cannot be converted to a slice of integers, it returns an error.
func (c *Conflex) GetIntSliceE(key string) ([]int, error) {
	return cast.ToIntSliceE(c.Get(key))
}

// GetStringSlice returns the value associated with the given key as a slice of strings.
// If the value is not found or cannot be converted to a slice of strings, an empty slice is returned.
func (c *Conflex) GetStringSlice(key string) []string {
	return cast.ToStringSlice(c.Get(key))
}

// GetStringSliceE returns the value associated with the given key as a slice of strings.
// If the value is not found or cannot be converted to a slice of strings, it returns an error.
func (c *Conflex) GetStringSliceE(key string) ([]string, error) {
	return cast.ToStringSliceE(c.Get(key))
}

// GetStringMap returns the value associated with the given key as a map[string]any.
// If the value is not found or cannot be converted to a map[string]any, the zero value is returned.
func (c *Conflex) GetStringMap(key string) map[string]any {
	return cast.ToStringMap(c.Get(key))
}

// GetStringMapE returns the value associated with the given key as a map[string]any.
// If the value is not found or cannot be converted to a map[string]any, it returns an error.
func (c *Conflex) GetStringMapE(key string) (map[string]any, error) {
	return cast.ToStringMapE(c.Get(key))
}

// GetStringMapString returns the value associated with the given key as a map[string]string.
// If the value is not found or cannot be converted to a map[string]string, the zero value is returned.
func (c *Conflex) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(c.Get(key))
}

// GetStringMapStringE returns the value associated with the given key as a map[string]string.
// If the value is not found or cannot be converted to a map[string]string, it returns an error.
func (c *Conflex) GetStringMapStringE(key string) (map[string]string, error) {
	return cast.ToStringMapStringE(c.Get(key))
}

// GetStringMapStringSlice returns the value associated with the given key as a map[string][]string.
// If the value is not found or cannot be converted to a map[string][]string, the zero value is returned.
func (c *Conflex) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(c.Get(key))
}

// GetStringMapStringSliceE returns the value associated with the given key as a map[string][]string.
// If the value is not found or cannot be converted to a map[string][]string, it returns an error.
func (c *Conflex) GetStringMapStringSliceE(key string) (map[string][]string, error) {
	return cast.ToStringMapStringSliceE(c.Get(key))
}
