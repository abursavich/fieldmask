// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"slices"
	"strings"
	"testing"

	"bursavich.dev/fieldmask/internal/testpb"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
)

var protoCmp = protocmp.Transform()

func protoDiff[T protoreflect.ProtoMessage](a, b T) string {
	return cmp.Diff(a, b, protoCmp)
}

func clone[T proto.Message](m T) T {
	return proto.Clone(m).(T)
}

func joinMasks(s ...string) string {
	return strings.Join(s, ",")
}

func simpleMsg(i int32, s string) *testpb.Message {
	return &testpb.Message{
		Int32Field:  i,
		StringField: s,
		MessageField: &testpb.Message{
			StringField: "nested-" + s,
		},
	}
}

func addGeneratedFields(m *testpb.Message) *testpb.Message {
	copyStringMapsToBytesMaps(m)
	copyStringMapKeysToLists(m)
	return m
}

func copyStringMapsToBytesMaps(m *testpb.Message) {
	m.MapBoolBytesField = stringMapToBytesMap(m.MapBoolStringField)
	m.MapStringBytesField = stringMapToBytesMap(m.MapStringStringField)
	m.MapInt32BytesField = stringMapToBytesMap(m.MapInt32StringField)
	m.MapInt64BytesField = stringMapToBytesMap(m.MapInt64StringField)
	m.MapSint32BytesField = stringMapToBytesMap(m.MapSint32StringField)
	m.MapSint64BytesField = stringMapToBytesMap(m.MapSint64StringField)
	m.MapUint32BytesField = stringMapToBytesMap(m.MapUint32StringField)
	m.MapUint64BytesField = stringMapToBytesMap(m.MapUint64StringField)
	m.MapFixed32BytesField = stringMapToBytesMap(m.MapFixed32StringField)
	m.MapFixed64BytesField = stringMapToBytesMap(m.MapFixed64StringField)
}

func stringMapToBytesMap[K comparable](m map[K]string) map[K][]byte {
	if m == nil {
		return nil
	}
	o := make(map[K][]byte, len(m))
	for k, v := range m {
		o[k] = []byte(v)
	}
	return o
}

func copyStringMapKeysToLists(m *testpb.Message) {
	m.RepeatedBoolField = mapKeysToList(m.MapBoolStringField, sortBool)
	m.RepeatedStringField = mapKeysToList(m.MapStringStringField, slices.Sort)
	m.RepeatedInt32Field = mapKeysToList(m.MapInt32StringField, slices.Sort)
	m.RepeatedInt64Field = mapKeysToList(m.MapInt64StringField, slices.Sort)
	m.RepeatedSint32Field = mapKeysToList(m.MapSint32StringField, slices.Sort)
	m.RepeatedSint64Field = mapKeysToList(m.MapSint64StringField, slices.Sort)
	m.RepeatedUint32Field = mapKeysToList(m.MapUint32StringField, slices.Sort)
	m.RepeatedUint64Field = mapKeysToList(m.MapUint64StringField, slices.Sort)
	m.RepeatedFixed32Field = mapKeysToList(m.MapFixed32StringField, slices.Sort)
	m.RepeatedFixed64Field = mapKeysToList(m.MapFixed64StringField, slices.Sort)
}

func mapKeysToList[K comparable](m map[K]string, sort func([]K)) []K {
	if m == nil {
		return nil
	}
	keys := maps.Keys(m)
	sort(keys)
	return keys
}

var testMsg = addGeneratedFields(&testpb.Message{
	BoolField:    true,
	StringField:  "2",
	Int32Field:   3,
	Int64Field:   4,
	Sint32Field:  5,
	Sint64Field:  6,
	Uint32Field:  7,
	Uint64Field:  8,
	Fixed32Field: 9,
	Fixed64Field: 10,
	MessageField: &testpb.Message{
		Int32Field:  11,
		StringField: "root",
		RepeatedStringField: []string{
			"nested-string(1)",
			"nested-string(2)",
		},
		RepeatedMessageField: []*testpb.Message{
			simpleMsg(0, "nested-repeated(0)"),
			simpleMsg(1, "nested-repeated(1)"),
		},
		MapStringStringField: map[string]string{
			"1": "nested-1",
			"2": "nested-2",
		},
		MapStringMessageField: map[string]*testpb.Message{
			"1": simpleMsg(0, "nested-string-map(0)"),
			"2": simpleMsg(0, "nested-string-map(0)"),
		},
	},
	BytesField: []byte("bytes"),

	RepeatedMessageField: []*testpb.Message{
		simpleMsg(0, "repeated(0)"),
		simpleMsg(1, "repeated(1)"),
		simpleMsg(2, "repeated(2)"),
		simpleMsg(3, "repeated(3)"),
	},
	RepeatedBytesField: [][]byte{
		[]byte("bytes(1)"),
		[]byte("bytes(2)"),
		[]byte("bytes(3)"),
	},

	MapBoolStringField: map[bool]string{
		false: "bool(false)",
		true:  "bool(true)",
	},
	MapStringStringField: map[string]string{
		"*":   "string(*)",
		"bar": "string(bar)",
		"foo": "string(foo)",
		"qux": "string(qux)",
	},
	MapInt32StringField: map[int32]string{
		-1: "int32(-1)",
		1:  "int32(1)",
		2:  "int32(2)",
		3:  "int32(3)",
	},
	MapInt64StringField: map[int64]string{
		-1: "int64(-1)",
		1:  "int64(1)",
		2:  "int64(2)",
		3:  "int64(3)",
	},
	MapSint32StringField: map[int32]string{
		-1: "sint32(-1)",
		1:  "sint32(1)",
		2:  "sint32(2)",
		3:  "sint32(3)",
	},
	MapSint64StringField: map[int64]string{
		-1: "sint64(-1)",
		1:  "sint64(1)",
		2:  "sint64(2)",
		3:  "sint64(3)",
	},
	MapUint32StringField: map[uint32]string{
		1: "uint32(1)",
		2: "uint32(2)",
		3: "uint32(3)",
	},
	MapUint64StringField: map[uint64]string{
		1: "uint64(1)",
		2: "uint64(2)",
		3: "uint64(3)",
	},
	MapFixed32StringField: map[uint32]string{
		1: "fixed32(1)",
		2: "fixed32(2)",
		3: "fixed32(3)",
	},
	MapFixed64StringField: map[uint64]string{
		1: "fixed64(1)",
		2: "fixed64(2)",
		3: "fixed64(3)",
	},

	MapBoolMessageField: map[bool]*testpb.Message{
		false: simpleMsg(0, "bool(false)"),
		true:  simpleMsg(1, "bool(true)"),
	},
	MapStringMessageField: map[string]*testpb.Message{
		"*":   simpleMsg(0, "string(*)"),
		"bar": simpleMsg(1, "string(bar)"),
		"foo": simpleMsg(2, "string(foo)"),
		"qux": simpleMsg(3, "string(qux)"),
	},
	MapInt32MessageField: map[int32]*testpb.Message{
		-1: simpleMsg(-1, "int32(-1)"),
		1:  simpleMsg(1, "int32(1)"),
		2:  simpleMsg(2, "int32(2)"),
		3:  simpleMsg(3, "int32(3)"),
	},
	MapInt64MessageField: map[int64]*testpb.Message{
		-1: simpleMsg(-1, "int64(-1)"),
		1:  simpleMsg(1, "int64(1)"),
		2:  simpleMsg(2, "int64(2)"),
		3:  simpleMsg(3, "int64(3)"),
	},
	MapSint32MessageField: map[int32]*testpb.Message{
		-1: simpleMsg(-1, "sint32(-1)"),
		1:  simpleMsg(1, "sint32(1)"),
		2:  simpleMsg(2, "sint32(2)"),
		3:  simpleMsg(3, "sint32(3)"),
	},
	MapSint64MessageField: map[int64]*testpb.Message{
		-1: simpleMsg(-1, "sint64(-1)"),
		1:  simpleMsg(1, "sint64(1)"),
		2:  simpleMsg(2, "sint64(2)"),
		3:  simpleMsg(3, "sint64(3)"),
	},
	MapUint32MessageField: map[uint32]*testpb.Message{
		1: simpleMsg(1, "uint32(1)"),
		2: simpleMsg(2, "uint32(2)"),
		3: simpleMsg(3, "uint32(3)"),
	},
	MapUint64MessageField: map[uint64]*testpb.Message{
		1: simpleMsg(1, "uint64(1)"),
		2: simpleMsg(2, "uint64(2)"),
		3: simpleMsg(3, "uint64(3)"),
	},
	MapFixed32MessageField: map[uint32]*testpb.Message{
		1: simpleMsg(1, "fixed32(1)"),
		2: simpleMsg(2, "fixed32(2)"),
		3: simpleMsg(3, "fixed32(3)"),
	},
	MapFixed64MessageField: map[uint64]*testpb.Message{
		1: simpleMsg(1, "fixed64(1)"),
		2: simpleMsg(2, "fixed64(2)"),
		3: simpleMsg(3, "fixed64(3)"),
	},
})

type basicTest struct {
	name  string
	mask  string
	opts  []Option
	err   bool
	paths []string
	msg   *testpb.Message
	out   *testpb.Message
}

func (tt basicTest) run(t *testing.T) {
	t.Helper()
	name := tt.name
	if name == "" {
		name = tt.mask
	}
	t.Run(name, func(t *testing.T) {
		t.Helper()
		fm, err := Parse[*testpb.Message](tt.mask, tt.opts...)
		if err != nil {
			if tt.err {
				return
			}
			t.Fatalf("Unexpected error parsing mask: %q: %v", tt.mask, err)
		}
		if tt.err {
			t.Fatalf("Expected error parsing mask: %q", tt.mask)
		}
		paths := fm.Paths()
		if diff := cmp.Diff(tt.paths, paths); diff != "" {
			t.Fatalf("Paths: unexpected diff:\n%s", diff)
		}
		masked := clone(tt.msg)
		fm.Mask(masked)
		if diff := protoDiff(tt.out, masked); diff != "" {
			t.Fatalf("Mask: unexpected diff:\n%s", diff)
		}
		output := fm.Clone(clone(tt.msg))
		if diff := protoDiff(tt.out, output); diff != "" {
			t.Fatalf("Clone: unexpected diff:\n%s", diff)
		}
	})
}

type updateTest struct {
	name string
	mask string
	opts []Option
	dst  *testpb.Message
	src  *testpb.Message
	out  *testpb.Message
	err  bool
}

func (tt updateTest) run(t *testing.T) {
	t.Helper()
	name := tt.name
	if name == "" {
		name = tt.mask
	}
	t.Run(name, func(t *testing.T) {
		t.Helper()
		fm, err := Parse[*testpb.Message](tt.mask, tt.opts...)
		if err != nil {
			t.Fatalf("Failed to parse mask: %q: %v", tt.mask, err)
		}
		dst := clone(tt.dst)
		if err := fm.Update(dst, tt.src); err != nil {
			if tt.err {
				return
			}
			t.Fatalf("Update: unexpected error: %v", err)
		}
		if tt.err {
			t.Fatalf("Update: expected error")
		}
		if diff := protoDiff(tt.out, dst); diff != "" {
			t.Fatalf("Update: unexpected diff:\n%s", diff)
		}
	})
}

type pathTest struct {
	name  string
	input string
	opts  []Option
	err   bool
	paths []string
}

func (tt pathTest) run(t *testing.T) {
	t.Helper()
	name := tt.name
	if name == "" {
		name = tt.input
	}
	t.Run(name, func(t *testing.T) {
		t.Helper()
		mask, err := Parse[*testpb.Message](tt.input, tt.opts...)
		if err != nil {
			if tt.err {
				return
			}
			t.Fatalf("Unexpected error: %v", err)
		}
		if tt.err {
			t.Fatal("Expected error")
		}
		paths := mask.Paths()
		if diff := cmp.Diff(tt.paths, paths); diff != "" {
			t.Fatalf("Unexpected paths diff:\n%s", diff)
		}
	})
}

func TestPaths(t *testing.T) {
	pathTest{
		name:  "asterisk",
		input: "*",
		paths: []string{"*"},
	}.run(t)

	pathTest{
		input: "int32_field",
		paths: []string{"int32_field"},
	}.run(t)

	pathTest{
		name:  "int32Field:json",
		input: "int32Field",
		opts:  []Option{WithFieldName(JSONFieldName, false)},
		paths: []string{"int32Field"},
	}.run(t)

	pathTest{
		name:  "int32Field:json-strict",
		input: "int32Field",
		opts:  []Option{WithFieldName(JSONFieldName, true)},
		paths: []string{"int32Field"},
	}.run(t)

	pathTest{
		name:  "int32_field:json",
		input: "int32_field",
		opts:  []Option{WithFieldName(JSONFieldName, false)},
		paths: []string{"int32Field"},
	}.run(t)

	pathTest{
		name:  "int32_field:json-strict",
		input: "int32_field",
		opts:  []Option{WithFieldName(JSONFieldName, true)},
		err:   true,
	}.run(t)

	pathTest{
		name:  "int32_field:text",
		input: "int32_field",
		opts:  []Option{WithFieldName(TextFieldName, false)},
		paths: []string{"int32_field"},
	}.run(t)

	pathTest{
		name:  "int32_field:text-strict",
		input: "int32_field",
		opts:  []Option{WithFieldName(TextFieldName, true)},
		paths: []string{"int32_field"},
	}.run(t)

	pathTest{
		name:  "int32Field:text",
		input: "int32Field",
		opts:  []Option{WithFieldName(TextFieldName, false)},
		paths: []string{"int32_field"},
	}.run(t)

	pathTest{
		name:  "int32Field:text-strict",
		input: "int32Field",
		opts:  []Option{WithFieldName(TextFieldName, true)},
		err:   true,
	}.run(t)

	pathTest{
		input: "string_field,int32_field",
		paths: []string{
			"int32_field",
			"string_field",
		},
	}.run(t)

	pathTest{
		input: "message_field",
		paths: []string{"message_field"},
	}.run(t)

	pathTest{
		input: "repeated_message_field",
		paths: []string{"repeated_message_field"},
	}.run(t)

	pathTest{
		input: "repeated_message_field.*",
		paths: []string{"repeated_message_field"},
	}.run(t)

	pathTest{
		input: joinMasks(
			"repeated_message_field.*.int32_field",
			"repeated_message_field.*.string_field",
		),
		paths: []string{
			"repeated_message_field.*.int32_field",
			"repeated_message_field.*.string_field",
		},
	}.run(t)

	pathTest{
		input: "map_string_message_field",
		paths: []string{"map_string_message_field"},
	}.run(t)

	pathTest{
		input: "map_string_message_field.*",
		paths: []string{"map_string_message_field"},
	}.run(t)

	pathTest{
		input: "map_string_message_field.foo",
		paths: []string{"map_string_message_field.foo"},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_string_message_field.foo.string_field",
			"map_string_message_field.foo.int32_field",
		),
		paths: []string{
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.foo.string_field",
		},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_string_message_field.foo",
			"map_string_message_field.bar",
		),
		paths: []string{
			"map_string_message_field.bar",
			"map_string_message_field.foo",
		},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.foo.string_field",
			"map_string_message_field.bar.string_field",
		),
		paths: []string{
			"map_string_message_field.bar.string_field",
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.foo.string_field",
		},
	}.run(t)

	pathTest{
		name: joinMasks(
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.foo.string_field",
			"map_string_message_field.bar.string_field",
		) + ":json-false",
		input: joinMasks(
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.foo.string_field",
			"map_string_message_field.bar.string_field",
		),
		opts: []Option{WithFieldName(JSONFieldName, false)},
		paths: []string{
			"mapStringMessageField.bar.stringField",
			"mapStringMessageField.foo.int32Field",
			"mapStringMessageField.foo.stringField",
		},
	}.run(t)

	pathTest{
		name: joinMasks(
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.foo.string_field",
			"map_string_message_field.bar.string_field",
		) + ":json-strict",
		input: joinMasks(
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.foo.string_field",
			"map_string_message_field.bar.string_field",
		),
		opts: []Option{WithFieldName(JSONFieldName, true)},
		err:  true,
	}.run(t)

	pathTest{
		name: joinMasks(
			"mapStringMessageField.foo.int32Field",
			"mapStringMessageField.foo.stringField",
			"mapStringMessageField.bar.stringField",
		) + ":json-strict",
		input: joinMasks(
			"mapStringMessageField.foo.int32Field",
			"mapStringMessageField.foo.stringField",
			"mapStringMessageField.bar.stringField",
		),
		opts: []Option{WithFieldName(JSONFieldName, true)},
		paths: []string{
			"mapStringMessageField.bar.stringField",
			"mapStringMessageField.foo.int32Field",
			"mapStringMessageField.foo.stringField",
		},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_string_message_field.bar",
			"map_string_message_field.foo.int32_field",
			"map_string_message_field", // wild after keys
		),
		paths: []string{
			"map_string_message_field",
		},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_string_message_field", // wild before keys
			"map_string_message_field.bar",
			"map_string_message_field.foo.int32_field",
		),
		paths: []string{
			"map_string_message_field",
		},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_string_message_field.bar.string_field",
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.*.int32_field", // wild after keys
		),
		paths: []string{
			"map_string_message_field.*.int32_field",
			"map_string_message_field.bar.string_field",
		},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_string_message_field.*.int32_field", // wild before keys
			"map_string_message_field.foo.int32_field",
			"map_string_message_field.bar.string_field",
		),
		paths: []string{
			"map_string_message_field.*.int32_field",
			"map_string_message_field.bar.string_field",
		},
	}.run(t)

	pathTest{
		input: joinMasks(
			"map_int32_message_field.11",
			"map_int32_message_field.2",
			"map_int32_message_field.-45",
		),
		paths: []string{
			"map_int32_message_field.-45",
			"map_int32_message_field.2",
			"map_int32_message_field.11",
		},
	}.run(t)
}
