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

package source

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/suite"
)

type ConsulSourceTestSuite struct {
	suite.Suite
	consul *Consul
	mockKV *mockKV
}

func (s *ConsulSourceTestSuite) SetupTest() {
	s.mockKV = &mockKV{}
	s.consul, _ = NewConsul("foo/bar", &mockDecoder{}, s.mockKV)
}

func TestConsulSourceTestSuite(t *testing.T) {
	suite.Run(t, new(ConsulSourceTestSuite))
}

func (s *ConsulSourceTestSuite) TestLoad_ValuePresent() {
	s.mockKV.pair = &api.KVPair{Key: "foo/bar", Value: []byte("value")}
	s.mockKV.meta = &api.QueryMeta{LastIndex: 123}
	s.consul.decoder = &mockDecoder{decodeMap: map[string]any{"foo": "bar"}}

	conf, err := s.consul.Load(context.Background())
	s.NoError(err)
	s.Equal(map[string]any{"foo": "bar"}, conf)
}

func (s *ConsulSourceTestSuite) TestLoad_ValueAbsent() {
	s.mockKV.pair = nil
	conf, err := s.consul.Load(context.Background())
	s.NoError(err)
	s.Empty(conf)
}

func (s *ConsulSourceTestSuite) TestLoad_DecodeError() {
	s.mockKV.pair = &api.KVPair{Key: "foo/bar", Value: []byte("bad")}
	s.mockKV.meta = &api.QueryMeta{LastIndex: 123}
	s.consul.decoder = &mockDecoder{err: errors.New("decode error")}
	_, err := s.consul.Load(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "decode error")
}

// --- Mocks ---

type mockKV struct {
	pair *api.KVPair
	meta *api.QueryMeta
}

func (m *mockKV) Get(_ string, _ *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	return m.pair, m.meta, nil
}

// mockDecoder implements codec.Decoder for testing

type mockDecoder struct {
	decodeMap map[string]any
	err       error
}

func (m *mockDecoder) Decode(_ []byte, v any) error {
	if m.err != nil {
		return m.err
	}
	if ptr, ok := v.(*map[string]any); ok {
		*ptr = m.decodeMap
		return nil
	}
	return errors.New("wrong type")
}
