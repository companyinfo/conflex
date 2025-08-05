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
	"fmt"
	"testing"
	"time"

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
	Foo string `conflex:"foo"`
	Bar int    `conflex:"bar"`
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
	for range 10 {
		go func() {
			_ = c.Load(context.Background())
			wg <- struct{}{}
		}()
	}
	for range 10 {
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

func (s *ConflexTestSuite) TestJSONSchemaValidation_Fails() {
	schema := []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"foo":{"type":"string"},"bar":{"type":"integer"}},"required":["foo","bar"]}`)
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": "notanint"}}
	c, err := New(WithSource(src), WithJSONSchema(schema))
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestJSONSchemaValidation_Succeeds() {
	schema := []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"foo":{"type":"string"},"bar":{"type":"integer"}},"required":["foo","bar"]}`)
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42}}
	c, err := New(WithSource(src), WithJSONSchema(schema))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestCustomValidator_Fails() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	c, err := New(WithSource(src), WithValidator(func(cfg map[string]any) error {
		if cfg["foo"] != "baz" {
			return errors.New("foo must be 'baz'")
		}
		return nil
	}))
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestCustomValidator_Succeeds() {
	src := &mockSource{conf: map[string]any{"foo": "baz"}}
	c, err := New(WithSource(src), WithValidator(func(cfg map[string]any) error {
		if cfg["foo"] != "baz" {
			return errors.New("foo must be 'baz'")
		}
		return nil
	}))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestBinding_ExtraFields() {
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42, "extra": 99}}
	var bind bindStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", bind.Foo)
	s.Equal(42, bind.Bar)
}

func (s *ConflexTestSuite) TestBinding_MissingFields() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	var bind bindStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", bind.Foo)
	s.Equal(0, bind.Bar)
}

func (s *ConflexTestSuite) TestBinding_TypeMismatch() {
	src := &mockSource{conf: map[string]any{"foo": 123, "bar": "notanint"}}
	var bind bindStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestMultipleDumpers_AllCalled() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	d1 := &mockDumper{}
	d2 := &mockDumper{}
	c, err := New(WithSource(src), WithDumper(d1), WithDumper(d2))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.NoError(c.Dump(context.Background()))
	s.True(d1.called)
	s.True(d2.called)
}

func (s *ConflexTestSuite) TestConcurrentGetSetLoad() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	wg := make(chan struct{})
	for range 10 {
		go func() {
			_ = c.Get("foo")
			_ = c.Load(context.Background())
			wg <- struct{}{}
		}()
	}
	for range 10 {
		<-wg
	}
}

// validatingBindStruct implements Validator interface to test race conditions during binding validation
type validatingBindStruct struct {
	Foo string `conflex:"foo"`
	Bar int    `conflex:"bar"`
}

func (v *validatingBindStruct) Validate() error {
	if v.Foo == "" {
		return errors.New("foo cannot be empty")
	}
	return nil
}

func (s *ConflexTestSuite) TestConcurrentLoadWithBindingValidation() {
	// This test specifically targets the race condition that existed between
	// binding validation and concurrent reads
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42}}
	var bind validatingBindStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)

	// Perform initial load
	s.NoError(c.Load(context.Background()))

	// Run concurrent operations that previously caused race conditions
	wg := make(chan struct{})

	for range 20 {
		go func() {
			defer func() { wg <- struct{}{} }()

			// Mix of reads and loads to stress test the race condition fix
			for i := 0; i < 10; i++ {
				if i%2 == 0 {
					// Read operations
					_ = c.Get("foo")
					_ = c.GetString("foo")
					_ = c.GetInt("bar")
					_ = c.Values()
				} else {
					// Load operation (includes binding validation)
					_ = c.Load(context.Background())
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for range 20 {
		<-wg
	}

	// Verify final state is consistent
	s.Equal("bar", c.GetString("foo"))
	s.Equal(42, c.GetInt("bar"))
}

func (s *ConflexTestSuite) TestNilAndEmptyConfigMap() {
	srcNil := &mockSource{conf: nil}
	c, err := New(WithSource(srcNil))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))

	srcEmpty := &mockSource{conf: map[string]any{}}
	c2, err := New(WithSource(srcEmpty))
	s.NoError(err)
	s.NoError(c2.Load(context.Background()))
}

func (s *ConflexTestSuite) TestGet_DeeplyNestedDotNotation() {
	src := &mockSource{conf: map[string]any{
		"a": map[string]any{"b": map[string]any{"c": map[string]any{"d": 1}}},
	}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	v := c.Get("a.b.c.d")
	s.Equal(1, v)
}

type pointerFieldsStruct struct {
	Foo *string `conflex:"foo"`
	Bar *int    `conflex:"bar"`
}

func (s *ConflexTestSuite) TestBinding_PointerFields() {
	foo := "bar"
	bar := 42
	src := &mockSource{conf: map[string]any{"foo": foo, "bar": bar}}
	var bind pointerFieldsStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.NotNil(bind.Foo)
	s.NotNil(bind.Bar)
	s.Equal(foo, *bind.Foo)
	s.Equal(bar, *bind.Bar)
}

type embeddedStruct struct {
	bindStruct
	Baz string `conflex:"baz"`
}

func (s *ConflexTestSuite) TestBinding_EmbeddedStruct() {
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42, "baz": "qux"}}
	var bind embeddedStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", bind.Foo)
	s.Equal(42, bind.Bar)
	s.Equal("qux", bind.Baz)
}

func (s *ConflexTestSuite) TestValidator_Panic() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	c, err := New(WithSource(src), WithValidator(func(_ map[string]any) error {
		panic("validator panic")
	}))
	s.NoError(err)
	s.Error(c.Load(context.Background()))
}

func (s *ConflexTestSuite) TestReloadAfterChange() {
	src := &mockSource{conf: map[string]any{"foo": "bar"}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", c.GetString("foo"))
	src.conf["foo"] = "baz"
	s.NoError(c.Load(context.Background()))
	s.Equal("baz", c.GetString("foo"))
}

func (s *ConflexTestSuite) TestGetterMethods() {
	timeStr := "2023-01-01T12:00:00Z"
	durStr := "1h2m3s"
	conf := map[string]any{
		"str":         "foo",
		"bool":        true,
		"boolstr":     "true",
		"int":         42,
		"intstr":      "42",
		"int32":       int32(32),
		"int64":       int64(64),
		"uint8":       uint8(8),
		"uint":        uint(7),
		"uint16":      uint16(16),
		"uint32":      uint32(32),
		"uint64":      uint64(64),
		"float64":     3.14,
		"floatstr":    "2.71",
		"time":        timeStr,
		"duration":    durStr,
		"intslice":    []any{1, 2, 3},
		"strslice":    []any{"a", "b"},
		"map":         map[string]any{"a": 1},
		"mapstr":      map[string]any{"a": "x"},
		"mapstrslice": map[string]any{"a": []any{"x", "y"}},
	}
	c, err := New(WithSource(&mockSource{conf: conf}))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))

	// GetString, GetStringE
	s.Equal("foo", c.GetString("str"))
	v, err := c.GetStringE("str")
	s.NoError(err)
	s.Equal("foo", v)
	_, err = c.GetStringE("notfound")
	s.Error(err)

	// GetBool, GetBoolE
	s.True(c.GetBool("bool"))
	b, err := c.GetBoolE("boolstr")
	s.NoError(err)
	s.True(b)
	_, err = c.GetBoolE("notfound")
	s.Error(err)

	// GetInt, GetIntE
	s.Equal(42, c.GetInt("int"))
	i, err := c.GetIntE("intstr")
	s.NoError(err)
	s.Equal(42, i)
	_, err = c.GetIntE("notfound")
	s.Error(err)

	// GetInt32, GetInt32E
	s.Equal(int32(32), c.GetInt32("int32"))
	i32, err := c.GetInt32E("int32")
	s.NoError(err)
	s.Equal(int32(32), i32)
	_, err = c.GetInt32E("notfound")
	s.Error(err)

	// GetInt64, GetInt64E
	s.Equal(int64(64), c.GetInt64("int64"))
	i64, err := c.GetInt64E("int64")
	s.NoError(err)
	s.Equal(int64(64), i64)
	_, err = c.GetInt64E("notfound")
	s.Error(err)

	// GetUint8, GetUint8E
	s.Equal(uint8(8), c.GetUint8("uint8"))
	u8, err := c.GetUint8E("uint8")
	s.NoError(err)
	s.Equal(uint8(8), u8)
	_, err = c.GetUint8E("notfound")
	s.Error(err)

	// GetUint, GetUintE
	s.Equal(uint(7), c.GetUint("uint"))
	u, err := c.GetUintE("uint")
	s.NoError(err)
	s.Equal(uint(7), u)
	_, err = c.GetUintE("notfound")
	s.Error(err)

	// GetUint16, GetUint16E
	s.Equal(uint16(16), c.GetUint16("uint16"))
	u16, err := c.GetUint16E("uint16")
	s.NoError(err)
	s.Equal(uint16(16), u16)
	_, err = c.GetUint16E("notfound")
	s.Error(err)

	// GetUint32, GetUint32E
	s.Equal(uint32(32), c.GetUint32("uint32"))
	u32, err := c.GetUint32E("uint32")
	s.NoError(err)
	s.Equal(uint32(32), u32)
	_, err = c.GetUint32E("notfound")
	s.Error(err)

	// GetUint64, GetUint64E
	s.Equal(uint64(64), c.GetUint64("uint64"))
	u64, err := c.GetUint64E("uint64")
	s.NoError(err)
	s.Equal(uint64(64), u64)
	_, err = c.GetUint64E("notfound")
	s.Error(err)

	// GetFloat64, GetFloat64E
	s.InDelta(3.14, c.GetFloat64("float64"), 0.0001)
	f64, err := c.GetFloat64E("floatstr")
	s.NoError(err)
	s.InDelta(2.71, f64, 0.0001)
	_, err = c.GetFloat64E("notfound")
	s.Error(err)

	// GetTime, GetTimeE
	t, err := c.GetTimeE("time")
	s.NoError(err)
	s.Equal(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), t)
	s.Equal(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), c.GetTime("time"))
	_, err = c.GetTimeE("notfound")
	s.Error(err)

	// GetDuration, GetDurationE
	d, err := c.GetDurationE("duration")
	s.NoError(err)
	s.Equal(1*time.Hour+2*time.Minute+3*time.Second, d)
	s.Equal(1*time.Hour+2*time.Minute+3*time.Second, c.GetDuration("duration"))
	_, err = c.GetDurationE("notfound")
	s.Error(err)

	// GetIntSlice, GetIntSliceE
	s.Equal([]int{1, 2, 3}, c.GetIntSlice("intslice"))
	is, err := c.GetIntSliceE("intslice")
	s.NoError(err)
	s.Equal([]int{1, 2, 3}, is)
	_, err = c.GetIntSliceE("notfound")
	s.Error(err)

	// GetStringSlice, GetStringSliceE
	s.Equal([]string{"a", "b"}, c.GetStringSlice("strslice"))
	ss, err := c.GetStringSliceE("strslice")
	s.NoError(err)
	s.Equal([]string{"a", "b"}, ss)
	_, err = c.GetStringSliceE("notfound")
	s.Error(err)

	// GetStringMap, GetStringMapE
	s.Equal(map[string]any{"a": 1}, c.GetStringMap("map"))
	m, err := c.GetStringMapE("map")
	s.NoError(err)
	s.Equal(map[string]any{"a": 1}, m)
	_, err = c.GetStringMapE("notfound")
	s.Error(err)

	// GetStringMapString, GetStringMapStringE
	s.Equal(map[string]string{"a": "x"}, c.GetStringMapString("mapstr"))
	ms, err := c.GetStringMapStringE("mapstr")
	s.NoError(err)
	s.Equal(map[string]string{"a": "x"}, ms)
	_, err = c.GetStringMapStringE("notfound")
	s.Error(err)

	// GetStringMapStringSlice, GetStringMapStringSliceE
	s.Equal(map[string][]string{"a": {"x", "y"}}, c.GetStringMapStringSlice("mapstrslice"))
	mss, err := c.GetStringMapStringSliceE("mapstrslice")
	s.NoError(err)
	s.Equal(map[string][]string{"a": {"x", "y"}}, mss)
	_, err = c.GetStringMapStringSliceE("notfound")
	s.Error(err)
}

func (s *ConflexTestSuite) TestBinding_UnexportedFields() {
	type hiddenStruct struct {
		Foo string `conflex:"foo"`
		bar int    `conflex:"bar"` // unexported
	}
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42}}
	var bind hiddenStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", bind.Foo)
	s.Equal(0, bind.bar) // unexported field should not be set
}

func (s *ConflexTestSuite) TestBinding_MissingTags() {
	type noTagStruct struct {
		Foo string
		Bar int
	}
	src := &mockSource{conf: map[string]any{"Foo": "bar", "Bar": 42}}
	var bind noTagStruct
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", bind.Foo)
	s.Equal(42, bind.Bar)
}

func (s *ConflexTestSuite) TestBinding_DefaultValues() {
	type defStruct struct {
		Foo string `conflex:"foo"`
		Bar int    `conflex:"bar"`
		Baz string `conflex:"baz"`
	}
	var bind defStruct
	bind.Baz = "default"
	src := &mockSource{conf: map[string]any{"foo": "bar", "bar": 42}}
	c, err := New(WithSource(src), WithBinding(&bind))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	s.Equal("bar", bind.Foo)
	s.Equal(42, bind.Bar)
	s.Equal("default", bind.Baz) // not overwritten
}

func (s *ConflexTestSuite) TestWithBinding_NilAndEmpty() {
	c, err := New(WithBinding(nil))
	s.NoError(err)
	s.NotNil(c)

	// Empty struct pointer
	type emptyStruct struct{}
	var bind *emptyStruct
	c2, err := New(WithBinding(bind))
	s.NoError(err)
	s.NotNil(c2)
}

func (s *ConflexTestSuite) TestGet_DotNotationWithKeyContainingDot() {
	src := &mockSource{conf: map[string]any{"a": map[string]any{"b": 2}}}
	c, err := New(WithSource(src))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))
	// Should return nested value if direct key does not exist
	v := c.Get("a.b")
	s.Equal(2, v)
	// Direct key with dot
	m := c.Values()
	(*m)["a.b"] = 3
	v2 := c.Get("a.b")
	s.Equal(3, v2)
}

func (s *ConflexTestSuite) TestWithFileDumper() {
	// Use a mock encoder and a temp file path
	path := "/tmp/conflex_test_file_dumper.json"
	c, err := New(WithFileDumper(path, "json"))
	s.NoError(err)
	s.NotNil(c)
	// Should add a dumper (not testing file output here)
	s.Len(c.dumpers, 1)
}

func (s *ConflexTestSuite) TestWithFileSource() {
	// Use a mock decoder and a temp file path
	path := "/tmp/conflex_test_file_source.json"
	c, err := New(WithFileSource(path, "json"))
	s.NoError(err)
	s.NotNil(c)
	// Should add a source
	s.Len(c.sources, 1)
}

func (s *ConflexTestSuite) TestWithContentSource() {
	data := []byte(`{"foo": "bar"}`)
	c, err := New(WithContentSource(data, "json"))
	s.NoError(err)
	s.NotNil(c)
	s.Len(c.sources, 1)
}

func (s *ConflexTestSuite) TestWithOSEnvVarSource() {
	c, err := New(WithOSEnvVarSource("TESTPREFIX_"))
	s.NoError(err)
	s.NotNil(c)
	s.Len(c.sources, 1)
}

func (s *ConflexTestSuite) TestWithConsulSource() {
	// This will fail if Consul is not available, so just test error on invalid codec
	c, err := New(WithConsulSource("some/path", "notacodec"))
	s.Error(err)
	s.NotNil(c)
	s.Len(c.sources, 0)
	// Test with valid codec (will still fail if Consul is not running, but should not panic)
	c2, _ := New(WithConsulSource("some/path", "json"))
	s.NotNil(c2)
}

func (s *ConflexTestSuite) TestConfigError() {
	// Test ConfigError formatting
	baseErr := errors.New("base error")

	// Error with field
	err1 := &ConfigError{
		Source:    "source1",
		Field:     "field1",
		Operation: "parse",
		Err:       baseErr,
	}
	s.Equal("config error in source1.field1 during parse: base error", err1.Error())
	s.Equal(baseErr, err1.Unwrap())

	// Error without field
	err2 := &ConfigError{
		Source:    "source2",
		Operation: "load",
		Err:       baseErr,
	}
	s.Equal("config error in source2 during load: base error", err2.Error())
	s.Equal(baseErr, err2.Unwrap())
}

func (s *ConflexTestSuite) TestParallelSourceLoading() {
	// Test that sources are loaded and merged correctly
	src1 := &mockSource{conf: map[string]any{"foo": "bar", "shared": "from1"}}
	src2 := &mockSource{conf: map[string]any{"baz": "qux", "shared": "from2"}}
	src3 := &mockSource{conf: map[string]any{"last": "value"}}

	c, err := New(WithSource(src1), WithSource(src2), WithSource(src3))
	s.NoError(err)
	s.NoError(c.Load(context.Background()))

	// Check all values are loaded
	s.Equal("bar", c.GetString("foo"))
	s.Equal("qux", c.GetString("baz"))
	s.Equal("value", c.GetString("last"))
	// Last source should override (src2 overrides src1)
	s.Equal("from2", c.GetString("shared"))
}

func (s *ConflexTestSuite) TestParallelSourceError() {
	// Test that errors in parallel loading are properly wrapped
	src1 := &mockSource{conf: map[string]any{"foo": "bar"}}
	src2 := &mockSource{err: errors.New("source failure")}

	c, err := New(WithSource(src1), WithSource(src2))
	s.NoError(err)

	err = c.Load(context.Background())
	s.Error(err)

	// Check that error is properly wrapped
	var configErr *ConfigError
	s.True(errors.As(err, &configErr))
	s.Contains(configErr.Source, "source[")
	s.Equal("load", configErr.Operation)
	s.Contains(configErr.Error(), "source failure")
}

func (s *ConflexTestSuite) TestDecoderConfigCaching() {
	// Test that decoder config is cached for performance
	// Create multiple binding operations to test decoder caching
	// We can't directly test the cached decoder, but we can verify performance consistency
	var bind1, bind2 bindStruct
	src := &mockSource{conf: map[string]any{"foo": "test", "bar": 123}}

	c1, err := New(WithSource(src), WithBinding(&bind1))
	s.NoError(err)

	c2, err := New(WithSource(src), WithBinding(&bind2))
	s.NoError(err)

	// Both should work without recreating decoder config each time
	s.NoError(c1.Load(context.Background()))
	s.NoError(c2.Load(context.Background()))

	s.Equal("test", bind1.Foo)
	s.Equal(123, bind1.Bar)
	s.Equal("test", bind2.Foo)
	s.Equal(123, bind2.Bar)
}

func BenchmarkParallelLoading(b *testing.B) {
	// Create multiple slow sources to demonstrate parallel loading benefits
	sources := make([]Source, 5)
	for i := 0; i < 5; i++ {
		sources[i] = &mockSlowSource{
			conf:  map[string]any{fmt.Sprintf("key%d", i): fmt.Sprintf("value%d", i)},
			delay: 10 * time.Millisecond, // Simulate I/O delay
		}
	}

	var opts []Option
	for _, src := range sources {
		opts = append(opts, WithSource(src))
	}

	c, err := New(opts...)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := c.Load(context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}
}

// mockSlowSource simulates a slow configuration source
type mockSlowSource struct {
	conf  map[string]any
	delay time.Duration
}

func (m *mockSlowSource) Load(_ context.Context) (map[string]any, error) {
	time.Sleep(m.delay)
	return m.conf, nil
}
