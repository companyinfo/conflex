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
	"testing"

	"github.com/stretchr/testify/suite"
)

// YAMLCodecTestSuite is a test suite for the YAMLCodec type.
type YAMLCodecTestSuite struct {
	suite.Suite
	codec YAMLCodec
}

// SetupTest sets up the test suite.
func (s *YAMLCodecTestSuite) SetupTest() {
	s.codec = YAMLCodec{}
}

// TestYAMLCodecTestSuite runs the YAMLCodecTestSuite.
func TestYAMLCodecTestSuite(t *testing.T) {
	suite.Run(t, new(YAMLCodecTestSuite))
}

// TestRegistration tests that the YAMLCodec is properly registered as both an
// encoder and decoder for the YAML data format.
func (s *YAMLCodecTestSuite) TestRegistration() {
	encoder, err := GetEncoder(TypeYAML)
	s.Require().NoError(err)
	s.Assert().IsType(YAMLCodec{}, encoder, "expected YAMLCodec, got %T", encoder)

	decoder, err := GetDecoder(TypeYAML)
	s.Require().NoError(err)
	s.Assert().IsType(YAMLCodec{}, decoder, "expected YAMLCodec, got %T", decoder)
}

func (s *YAMLCodecTestSuite) TestEncode() {
	data := map[string]any{"foo": "bar", "num": 42, "nested": map[string]any{"key": "value"}}
	b, err := s.codec.Encode(data)
	s.NoError(err)
	// Basic check, YAML output can vary (e.g. order of keys)
	s.Contains(string(b), "foo: bar")
	s.Contains(string(b), "num: 42")
	s.Contains(string(b), "nested:")
	s.Contains(string(b), "  key: value")
}

func (s *YAMLCodecTestSuite) TestEncode_Empty() {
	b, err := s.codec.Encode(map[string]any{})
	s.NoError(err)
	// gopkg.in/yaml.v3 marshals an empty map to "null\n" or "{}\n" depending on context
	// For consistency, we'll accept either, or just check for non-error and minimal output.
	// "{}\n" is a common representation. "null\n" can also occur.
	// Let's check if it's one of the expected empty representations.
	strOut := string(b)
	s.True(strOut == "{}\n" || strOut == "null\n" || strOut == "")
}

func (s *YAMLCodecTestSuite) TestEncode_Error() {
	// Channels are not directly serializable to YAML by default by gopkg.in/yaml.v3
	ch := make(chan int)
	_, err := s.codec.Encode(ch)
	s.Error(err)
}

func (s *YAMLCodecTestSuite) TestDecode() {
	var v map[string]any
	yamlStr := `
foo: bar
num: 42
nested:
  key: value
`
	err := s.codec.Decode([]byte(yamlStr), &v)
	s.NoError(err)
	s.Equal("bar", v["foo"])
	s.EqualValues(42, v["num"])
	s.Require().IsType(map[string]any{}, v["nested"])
	nestedMap := v["nested"].(map[string]any)
	s.Equal("value", nestedMap["key"])
}

func (s *YAMLCodecTestSuite) TestDecode_Empty() {
	var v map[string]any
	err := s.codec.Decode([]byte(`{}`), &v)
	s.NoError(err)
	s.Empty(v)

	var v2 map[string]any
	err = s.codec.Decode([]byte(`null`), &v2)
	s.NoError(err)
	s.Nil(v2) // Unmarshalling "null" into a map results in a nil map
}

func (s *YAMLCodecTestSuite) TestDecode_Error() {
	var v map[string]any
	// Use a truly invalid YAML: unclosed quote
	err := s.codec.Decode([]byte("foo: \"bar"), &v)
	s.Error(err)
}

func (s *YAMLCodecTestSuite) TestDecode_IntoStruct() {
	type Config struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	var cfg Config
	yamlStr := `
host: localhost
port: 8080
`
	err := s.codec.Decode([]byte(yamlStr), &cfg)
	s.NoError(err)
	s.Equal("localhost", cfg.Host)
	s.Equal(8080, cfg.Port)
}
