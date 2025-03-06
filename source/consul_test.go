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
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ConsulTestSuite is a test suite for testing the Consul configuration source.
type ConsulTestSuite struct {
	suite.Suite
}

// TestNewConsulInvalidAddress tests that NewConsul returns an error when the provided Consul address is invalid.
func (s *ConsulTestSuite) TestNewConsulInvalidAddress() {
	consul, err := NewConsul(&api.Config{Address: "invalid://address"}, "test/path", nil)
	s.Require().Error(err)
	s.Require().Nil(consul)

	s.Assert().Contains(err.Error(), "failed to create consul client")
}

// TestConsulSuite runs the test suite for the Consul configuration source.
func TestConsulSuite(t *testing.T) {
	suite.Run(t, new(ConsulTestSuite))
}
