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

package codec

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type EnvVarCodecTestSuite struct {
	suite.Suite
	codec EnvVarCodec
}

func (s *EnvVarCodecTestSuite) SetupTest() {
	s.codec = EnvVarCodec{}
}

func TestEnvVarCodecTestSuite(t *testing.T) {
	suite.Run(t, new(EnvVarCodecTestSuite))
}

func (s *EnvVarCodecTestSuite) TestDecode_Simple() {
	data := []byte("FOO=bar\nBAZ=qux")
	var v map[string]any
	err := s.codec.Decode(data, &v)
	s.NoError(err)
	s.Equal("bar", v["foo"])
	s.Equal("qux", v["baz"])
}

func (s *EnvVarCodecTestSuite) TestDecode_Nested() {
	data := []byte("DATABASE_HOST=localhost\nDATABASE_PORT=5432\nDATABASE_USER_NAME=admin")
	var v map[string]any
	err := s.codec.Decode(data, &v)
	s.NoError(err)
	db, ok := v["database"].(map[string]any)
	s.True(ok)
	s.Equal("localhost", db["host"])
	s.Equal("5432", db["port"])
	user, ok := db["user"].(map[string]any)
	s.True(ok)
	s.Equal("admin", user["name"])
}

func (s *EnvVarCodecTestSuite) TestDecode_Empty() {
	data := []byte("")
	var v map[string]any
	err := s.codec.Decode(data, &v)
	s.NoError(err)
	s.Empty(v)
}

func (s *EnvVarCodecTestSuite) TestDecode_Malformed() {
	data := []byte("FOO\nBAR=baz") // FOO has no '='
	var v map[string]any
	err := s.codec.Decode(data, &v)
	s.NoError(err)
	s.Equal("baz", v["bar"])
	s.NotContains(v, "foo")
}

func (s *EnvVarCodecTestSuite) TestDecode_WrongType() {
	data := []byte("FOO=bar")
	var v []string // not a *map[string]any
	err := s.codec.Decode(data, &v)
	s.Error(err)
}
