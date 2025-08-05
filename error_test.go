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

package conflex

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ErrorTestSuite struct {
	suite.Suite
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

func (s *ErrorTestSuite) TestConfigError() {
	baseErr := errors.New("base error")

	// Test direct struct creation (backward compatibility)
	err1 := &ConfigError{
		Source:    "source1",
		Field:     "field1",
		Operation: "parse",
		Err:       baseErr,
	}
	s.Equal("config error in source1.field1 during parse: base error", err1.Error())
	s.Equal(baseErr, err1.Unwrap())

	// Test without field
	err2 := &ConfigError{
		Source:    "source2",
		Operation: "load",
		Err:       baseErr,
	}
	s.Equal("config error in source2 during load: base error", err2.Error())
	s.Equal(baseErr, err2.Unwrap())
}

func (s *ErrorTestSuite) TestNewConfigError() {
	baseErr := errors.New("test error")

	err := NewConfigError("test-source", "test-operation", baseErr)

	s.NotNil(err)
	s.Equal("test-source", err.Source)
	s.Equal("test-operation", err.Operation)
	s.Equal("", err.Field) // Should be empty when using NewConfigError
	s.Equal(baseErr, err.Err)
	s.Equal(baseErr, err.Unwrap())
	s.Equal("config error in test-source during test-operation: test error", err.Error())
}

func (s *ErrorTestSuite) TestNewConfigFieldError() {
	baseErr := errors.New("field error")

	err := NewConfigFieldError("test-source", "test-field", "test-operation", baseErr)

	s.NotNil(err)
	s.Equal("test-source", err.Source)
	s.Equal("test-field", err.Field)
	s.Equal("test-operation", err.Operation)
	s.Equal(baseErr, err.Err)
	s.Equal(baseErr, err.Unwrap())
	s.Equal("config error in test-source.test-field during test-operation: field error", err.Error())
}

func (s *ErrorTestSuite) TestConfigErrorErrorWrapping() {
	// Test that ConfigError properly supports error wrapping/unwrapping
	originalErr := errors.New("original error")
	configErr := NewConfigError("source", "operation", originalErr)

	// Test errors.Is
	s.True(errors.Is(configErr, originalErr))

	// Test errors.As
	var targetErr *ConfigError
	s.True(errors.As(configErr, &targetErr))
	s.Equal("source", targetErr.Source)
	s.Equal("operation", targetErr.Operation)
}

func (s *ErrorTestSuite) TestConfigErrorChaining() {
	// Test chaining multiple ConfigErrors
	originalErr := errors.New("root cause")
	firstErr := NewConfigError("first-source", "first-op", originalErr)
	secondErr := NewConfigError("second-source", "second-op", firstErr)

	// Should be able to unwrap to the original error
	s.True(errors.Is(secondErr, originalErr))
	s.True(errors.Is(secondErr, firstErr))

	// Test the error message of the outer error
	s.Contains(secondErr.Error(), "second-source")
	s.Contains(secondErr.Error(), "second-op")
}
