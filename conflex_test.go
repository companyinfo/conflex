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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConflexTestSuite struct {
	suite.Suite
}

func TestConflexTestSuite(t *testing.T) {
	suite.Run(t, new(ConflexTestSuite))
}

type mockSource struct {
	conf map[string]any
	err  error
}

func (m *mockSource) Load(_ context.Context) (map[string]any, error) {
	return m.conf, m.err
}

type mockDumper struct {
	called bool
	values *map[string]any
	err    error
}

func (m *mockDumper) Dump(_ context.Context, values *map[string]any) error {
	m.called = true
	m.values = values
	return m.err
}

type bindStruct struct {
	Foo string `config:"foo"`
	Bar int    `config:"bar"`
}

func (s *ConflexTestSuite) TestNewAndLoad_Success() {
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NotNil(c)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", c.GetString("foo"))
	s.Equal(42, c.GetInt("bar"))
}

func (s *ConflexTestSuite) TestNew_MultipleSources_Merge() {
	src1 := &mockSource{conf: map[string]any{"foo": "bar", "bar": 1}}
	src2 := &mockSource{conf: map[string]any{"bar": 2, "baz": 3}}
	c, err := New(WithSource(src1), WithSource(src2))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", c.GetString("foo"))
	s.Equal(2, c.GetInt("bar")) // src2 overrides src1
	s.Equal(3, c.GetInt("baz"))
}

func (s *ConflexTestSuite) TestWithBinding() {
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42}}
	var bind bindStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", bind.Foo)
	s.Equal(42, bind.Bar)
}

func (s *ConflexTestSuite) TestLoad_ErrorPropagates() {
	src := &mockSource{err: errors.New("fail")}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestDump_CallsDumper() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	dumper := &mockDumper{}
	c, err := New(WithSource(src), WithDumper(dumper))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.NoError(c.Dump(context.Background()))
	s.True(dumper.called)
	s.NotNil(dumper.values)
	s.Equal("bar", (*dumper.values)["foo"])
}

func (s *ConflexTestSuite) TestGet_NotFound() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Nil(c.Get("notfound"))
	s.Equal("", c.GetString("notfound"))
	s.Equal(0, c.GetInt("notfound"))
}

func (s *ConflexTestSuite) TestWithFileSource_ErrorDecoder() {
	badDecoder := &mockDecoder{err: errors.New("decode error")}
	c, err := New(WithSource(&mockFileSource{decoder: badDecoder}))
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

type mockDecoder struct {
	err error
}

func (m *mockDecoder) Decode(_ []byte, _ any) error {
	return m.err
}

type mockFileSource struct {
	decoder *mockDecoder
}

func (m *mockFileSource) Load(_ context.Context) (map[string]any, error) {
	var v map[string]any
	err := m.decoder.Decode([]byte("bad"), &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *ConflexTestSuite) TestWithContentSource_ErrorDecoder() {
	badDecoder := &mockDecoder{err: errors.New("decode error")}
	c, err := New(WithSource(&mockFileSource{decoder: badDecoder}))
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestDump_ErrorPropagates() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	dumper := &mockDumper{err: errors.New("dump error")}
	c, err := New(WithSource(src), WithDumper(dumper))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Error(c.Dump(context.Background()))
}

func (s *ConflexTestSuite) TestWithBinding_NonPointer() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	var bind bindStruct
	c, err := New(WithSource(src), WithBinding(bind)) // not a pointer
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestLoad_NoSources() {
	c, err := New()
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestDump_NoDumpers() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.NoError(c.Dump(context.Background()))
}

func (s *ConflexTestSuite) TestGetStringMap_And_GetStringSlice() {
	src := &mockSource{conf: map[string]any{"m": map[string]any{"a": 1}, "s": []any{"a", "b"}}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	m := c.GetStringMap("m")
	s.Equal(map[string]any{"a": 1}, m)
	slice := c.GetStringSlice("s")
	s.Equal([]string{"a", "b"}, slice)
}

func (s *ConflexTestSuite) TestConcurrentLoad() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	c, err := New(WithSource(src))
	s.NoError(err)
	wg := make(chan struct{})
	for range [10]int{} {
		go func() {
			_ = c.Load(context.Background())
			wg <- struct{}{}
		}()
	}
	for range [10]int{} {
		<-wg
	}
}

func (s *ConflexTestSuite) TestWithSource_Nil() {
	c, err := New(WithSource(nil))
	s.NoError(err)
	s.NotNil(c)
}

func (s *ConflexTestSuite) TestWithDumper_Nil() {
	c, err := New(WithDumper(nil))
	s.NoError(err)
	s.NotNil(c)
}

func (s *ConflexTestSuite) TestWithBinding_Nil() {
	c, err := New(WithBinding(nil))
	s.NoError(err)
	s.NotNil(c)
}

func (s *ConflexTestSuite) TestNew_NoOptions() {
	c, err := New()
	s.NoError(err)
	s.NotNil(c)
}

func (s *ConflexTestSuite) TestLoad_SourceReturnsNilMap() {
	src := &mockSource{conf: nil}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestDump_DumperReturnsError() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	dumper := &mockDumper{err: errors.New("fail dump")}
	c, err := New(WithSource(src), WithDumper(dumper))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Error(c.Dump(context.Background()))
}

func (s *ConflexTestSuite) TestGetStringMapString_And_GetStringMapStringSlice() {
	src := &mockSource{conf: map[string]any{
		"m":  map[string]any{"a": "x", "b": "y"},
		"ms": map[string]any{"a": []any{"x", "y"}},
	}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	m := c.GetStringMapString("m")
	s.Equal(map[string]string{"a": "x", "b": "y"}, m)
	ms := c.GetStringMapStringSlice("ms")
	s.Equal(map[string][]string{"a": {"x", "y"}}, ms)
}

func (s *ConflexTestSuite) TestGetIntSlice_NonIntValues() {
	src := &mockSource{conf: map[string]any{"ints": []any{"1", "2", 3}}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	slice := c.GetIntSlice("ints")
	s.Equal([]int{1, 2, 3}, slice)
}

func (s *ConflexTestSuite) TestGetStringSlice_NonStringValues() {
	src := &mockSource{conf: map[string]any{"strs": []any{1, 2, "a"}}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	slice := c.GetStringSlice("strs")
	s.Equal([]string{"1", "2", "a"}, slice)
}

func (s *ConflexTestSuite) TestGet_NestedDotNotation() {
	src := &mockSource{conf: map[string]any{
		"outer": map[string]any{
			"inner": map[string]any{
				"val": 42,
			},
		},
	}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	v := c.Get("outer.inner.val")
	s.Equal(42, v)
}
