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
	"github.com/stretchr/testify/suite"
	"testing"
)

// YAMLCodecTestSuite is a test suite for the YAMLCodec type.
type YAMLCodecTestSuite struct {
	suite.Suite
}

// TestRegistration tests that the YAMLCodec is properly registered as both an
// encoder and decoder for the YAML data format.
func (s *YAMLCodecTestSuite) TestRegistration() {
	encoder, err := GetEncoder(TypeYAML)
	s.Require().NoError(err)
	s.Assert().Equal(YAMLCodec{}, encoder, "expected YAMLCodec, got %v", encoder)

	decoder, err := GetDecoder(TypeYAML)
	s.Require().NoError(err)
	s.Assert().Equal(YAMLCodec{}, decoder, "expected YAMLCodec, got %v", decoder)
}

// TestYAMLCodec runs the test suite for the YAMLCodecTestSuite.
func TestYAMLCodec(t *testing.T) {
	suite.Run(t, new(YAMLCodecTestSuite))
}
