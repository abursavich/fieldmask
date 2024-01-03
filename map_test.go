// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"bursavich.dev/fieldmask/internal/testpb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestBoolMap(t *testing.T) {
	testMapType[bool](t, "bool", sortBool)
}

func TestStringMap(t *testing.T) {
	testMapType[string](t, "string", slices.Sort)
}

func TestInt32Map(t *testing.T) {
	testMapType[int32](t, "int32", slices.Sort)
}

func TestInt64Map(t *testing.T) {
	testMapType[int64](t, "int64", slices.Sort)
}

func TestSint32Map(t *testing.T) {
	testMapType[int32](t, "sint32", slices.Sort)
}

func TestSint64Map(t *testing.T) {
	testMapType[int64](t, "sint64", slices.Sort)
}

func TestUint32Map(t *testing.T) {
	testMapType[uint32](t, "uint32", slices.Sort)
}

func TestUint64Map(t *testing.T) {
	testMapType[uint64](t, "uint64", slices.Sort)
}

func TestFixed32Map(t *testing.T) {
	testMapType[uint32](t, "fixed32", slices.Sort)
}

func TestFixed64Map(t *testing.T) {
	testMapType[uint64](t, "fixed64", slices.Sort)
}

func testMapType[K comparable](t *testing.T, keyType string, sort func([]K)) {
	t.Run("string", func(t *testing.T) {
		tt := newMapTest[K](keyType, "string", sort, func(v protoreflect.Value) protoreflect.Value {
			s := v.String()
			s = s + "/" + s
			return protoreflect.ValueOfString(s)
		})
		tt.runBasic(t)
		tt.runUpdate(t)
	})
	t.Run("bytes", func(t *testing.T) {
		tt := newMapTest[K](keyType, "bytes", sort, func(v protoreflect.Value) protoreflect.Value {
			b := v.Bytes()
			b = bytes.Join([][]byte{b, b}, []byte("/"))
			return protoreflect.ValueOfBytes(b)
		})
		tt.runBasic(t)
		tt.runUpdate(t)
	})
	t.Run("message", func(t *testing.T) {
		tt := newMapTest[K](keyType, "message", sort, func(v protoreflect.Value) protoreflect.Value {
			m := clone(v.Message().Interface().(*testpb.Message))
			m.Int32Field *= 2
			m.StringField = m.StringField + "/" + m.StringField
			return v
		})
		tt.runBasic(t)
		tt.runUpdate(t)
		tt.runMessage(t)
	})
}

type mapTest[K comparable] struct {
	field   string
	fd      protoreflect.FieldDescriptor
	keys    []K
	paths   []string
	mapOnly *testpb.Message
	mutate  func(protoreflect.Value) protoreflect.Value
}

func newMapTest[K comparable](keyType, valueType string, sort func([]K), mutate func(protoreflect.Value) protoreflect.Value) *mapTest[K] {
	field := "map_" + keyType + "_" + valueType + "_field"
	fd := testMsg.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(field))
	var keys []K
	mapOnly := &testpb.Message{}
	testMsg.ProtoReflect().Get(fd).Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		keys = append(keys, k.Interface().(K))
		mapOnly.ProtoReflect().Mutable(fd).Map().Set(k, v)
		return true
	})
	sort(keys)
	paths := make([]string, len(keys))
	for i, k := range keys {
		paths[i] = field + "." + maybeQuote(fmt.Sprint(k))
	}
	return &mapTest[K]{
		field:   field,
		fd:      fd,
		keys:    keys,
		paths:   paths,
		mapOnly: mapOnly,
		mutate:  mutate,
	}
}

func (tt *mapTest[K]) get(m *testpb.Message, key K) protoreflect.Value {
	mk := protoreflect.MapKey(protoreflect.ValueOf(key))
	return m.ProtoReflect().Get(tt.fd).Map().Get(mk)
}

func (tt *mapTest[K]) set(m *testpb.Message, key K, val protoreflect.Value) {
	mk := protoreflect.MapKey(protoreflect.ValueOf(key))
	m.ProtoReflect().Mutable(tt.fd).Map().Set(mk, val)
}

func (tt *mapTest[K]) delete(m *testpb.Message, key K) {
	mk := protoreflect.MapKey(protoreflect.ValueOf(key))
	m.ProtoReflect().Mutable(tt.fd).Map().Clear(mk)
}

func (tt *mapTest[K]) copy(dst, src *testpb.Message, key K) {
	tt.set(dst, key, tt.get(src, key))
}

func (tt *mapTest[K]) copyMutated(dst, src *testpb.Message, key K) {
	tt.set(dst, key, tt.mutate(tt.get(src, key)))
}

func (tt *mapTest[K]) runBasic(t *testing.T) {
	if reflect.ValueOf(tt.keys[0]).Kind() != reflect.String {
		basicTest{
			mask: tt.field + ".invalid_key",
			err:  true,
		}.run(t)
	}

	basicTest{
		mask: tt.field + ".*.invalid_subfield",
		err:  true,
	}.run(t)

	basicTest{
		mask: tt.field + ".foo.invalid_subfield",
		err:  true,
	}.run(t)

	basicTest{
		mask:  tt.field,
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.mapOnly,
	}.run(t)

	basicTest{
		mask:  tt.field + ".*",
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.mapOnly,
	}.run(t)

	basicTest{
		mask:  tt.paths[0],
		paths: []string{tt.paths[0]},
		msg:   testMsg,
		out: func() *testpb.Message {
			var dst testpb.Message
			tt.set(&dst, tt.keys[0], tt.get(testMsg, tt.keys[0]))
			return &dst
		}(),
	}.run(t)

	basicTest{
		mask: joinMasks(
			tt.paths[1],
			tt.paths[0],
		),
		paths: []string{
			tt.paths[0],
			tt.paths[1],
		},
		msg: testMsg,
		out: func() *testpb.Message {
			var dst testpb.Message
			tt.copy(&dst, testMsg, tt.keys[0])
			tt.copy(&dst, testMsg, tt.keys[1])
			return &dst
		}(),
	}.run(t)

	basicTest{
		mask: joinMasks(
			tt.field+".*", // wild before keys
			tt.paths[0],
			tt.paths[1],
		),
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.mapOnly,
	}.run(t)

	basicTest{
		mask: joinMasks(
			tt.paths[0],
			tt.field+".*", // wild between keys
			tt.paths[1],
		),
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.mapOnly,
	}.run(t)

	basicTest{
		mask: joinMasks(
			tt.paths[0],
			tt.paths[1],
			tt.field+".*", // wild after keys
		),
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.mapOnly,
	}.run(t)
}

func (tt *mapTest[K]) runUpdate(t *testing.T) {
	updateTest{
		name: tt.field + ":no-op",
		mask: tt.field,
		dst:  testMsg,
		src:  testMsg,
		out:  testMsg,
	}.run(t)

	updateTest{
		name: tt.field + ":src-nil",
		mask: tt.field,
		dst:  testMsg,
		src:  nil,
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.field + ":src-map-nil",
		mask: tt.field,
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.field + ":dst-map-nil",
		mask: tt.field,
		dst: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		src: testMsg,
		out: testMsg,
	}.run(t)

	updateTest{
		name: tt.field + ":both-maps-nil",
		mask: tt.field,
		dst: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.paths[0] + ":src-map-nil",
		mask: tt.paths[0],
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			tt.delete(out, tt.keys[0])
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.paths[0] + ":dst-map-nil",
		mask: tt.paths[0],
		dst: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		src: testMsg,
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			tt.copy(out, testMsg, tt.keys[0])
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.paths[0] + ":both-maps-nil",
		mask: tt.paths[0],
		dst: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.paths[0] + ":key-missing",
		mask: tt.paths[0],
		dst: func() *testpb.Message {
			out := clone(testMsg)
			tt.delete(out, tt.keys[0])
			return out
		}(),
		src: func() *testpb.Message {
			out := clone(testMsg)
			tt.delete(out, tt.keys[0])
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			tt.delete(out, tt.keys[0])
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.paths[0] + ":src-key-missing",
		mask: tt.paths[0],
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			tt.delete(out, tt.keys[0])
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			tt.delete(out, tt.keys[0])
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.paths[0] + ":dst-key-missing",
		mask: tt.paths[0],
		dst: func() *testpb.Message {
			out := clone(testMsg)
			tt.delete(out, tt.keys[0])
			return out
		}(),
		src: testMsg,
		out: testMsg,
	}.run(t)

	updateTest{
		name: tt.paths[0] + ":mutated",
		mask: tt.paths[0],
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			for _, key := range tt.keys {
				tt.copyMutated(out, testMsg, key)
			}
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			tt.copyMutated(out, testMsg, tt.keys[0])
			return out
		}(),
	}.run(t)

	updateTest{
		mask: joinMasks(
			tt.paths[0],
			tt.paths[1],
		),
		dst: testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			tt.copyMutated(out, testMsg, tt.keys[0])
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			// mutate tt.keys[0]
			tt.copyMutated(out, testMsg, tt.keys[0])
			// skip tt.keys[1]
			for _, key := range tt.keys[2:] {
				// keep tt.keys[2:]
				tt.copy(out, testMsg, key)
			}
			return out
		}(),
	}.run(t)
}

func (tt *mapTest[K]) runMessage(t *testing.T) {
	updateTest{
		mask: tt.field + ".*.int32_field",
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Get(tt.fd).Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				m := v.Message().Interface().(*testpb.Message)
				m.Int32Field *= 2
				m.StringField = "ignore"
				m.MessageField = &testpb.Message{StringField: "ignore-nested", Int32Field: 42}
				return true
			})
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Get(tt.fd).Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				m := v.Message().Interface().(*testpb.Message)
				m.Int32Field *= 2
				return true
			})
			return out
		}(),
	}.run(t)

	updateTest{
		mask: joinMasks(
			tt.field+".*.int32_field",
			tt.field+".*.string_field",
		),
		dst: testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Get(tt.fd).Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				m := v.Message().Interface().(*testpb.Message)
				m.Int32Field *= 2
				m.StringField = m.StringField + "/" + m.StringField
				return true
			})
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Get(tt.fd).Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				m := v.Message().Interface().(*testpb.Message)
				m.StringField = m.StringField + "/" + m.StringField
				m.Int32Field *= 2
				return true
			})
			return out
		}(),
	}.run(t)

	updateTest{
		mask: joinMasks(
			tt.field+".*.int32_field",
			tt.paths[0]+".string_field",
		),
		dst: testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Get(tt.fd).Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				m := v.Message().Interface().(*testpb.Message)
				m.Int32Field *= 2
				m.StringField = m.StringField + "/" + m.StringField
				return true
			})
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			mm := out.ProtoReflect().Get(tt.fd).Map()
			mm.Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				m := v.Message().Interface().(*testpb.Message)
				m.Int32Field *= 2
				return true
			})
			m := mm.Get(protoreflect.MapKey(protoreflect.ValueOf(tt.keys[0]))).Message().Interface().(*testpb.Message)
			m.StringField = m.StringField + "/" + m.StringField
			return out
		}(),
	}.run(t)
}

func sortBool(s []bool) {
	slices.SortFunc(s, func(a, b bool) int {
		if !a && b {
			return -1
		}
		if a == b {
			return 0
		}
		return 1
	})
}
