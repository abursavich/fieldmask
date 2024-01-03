// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"bytes"
	"testing"

	"bursavich.dev/fieldmask/internal/testpb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestBoolList(t *testing.T) {
	tt := newListTest("bool", func(v protoreflect.Value) protoreflect.Value {
		x := v.Bool()
		return protoreflect.ValueOfBool(!x)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestStringList(t *testing.T) {
	tt := newListTest("string", func(v protoreflect.Value) protoreflect.Value {
		x := v.String()
		return protoreflect.ValueOfString(x + "/" + x)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestInt32List(t *testing.T) {
	tt := newListTest("int32", func(v protoreflect.Value) protoreflect.Value {
		x := int32(v.Int())
		return protoreflect.ValueOfInt32(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestInt64List(t *testing.T) {
	tt := newListTest("int64", func(v protoreflect.Value) protoreflect.Value {
		x := v.Int()
		return protoreflect.ValueOfInt64(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestSint32List(t *testing.T) {
	tt := newListTest("sint32", func(v protoreflect.Value) protoreflect.Value {
		x := int32(v.Int())
		return protoreflect.ValueOfInt32(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestSint64List(t *testing.T) {
	tt := newListTest("sint64", func(v protoreflect.Value) protoreflect.Value {
		x := v.Int()
		return protoreflect.ValueOfInt64(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestUint32List(t *testing.T) {
	tt := newListTest("uint32", func(v protoreflect.Value) protoreflect.Value {
		x := uint32(v.Uint())
		return protoreflect.ValueOfUint32(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestUint64List(t *testing.T) {
	tt := newListTest("uint64", func(v protoreflect.Value) protoreflect.Value {
		x := v.Uint()
		return protoreflect.ValueOfUint64(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestFixed32List(t *testing.T) {
	tt := newListTest("fixed32", func(v protoreflect.Value) protoreflect.Value {
		x := uint32(v.Uint())
		return protoreflect.ValueOfUint32(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestFixed64List(t *testing.T) {
	tt := newListTest("fixed64", func(v protoreflect.Value) protoreflect.Value {
		x := v.Uint()
		return protoreflect.ValueOfUint64(x * 2)
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

func TestMessageList(t *testing.T) {
	tt := newListTest("message", func(v protoreflect.Value) protoreflect.Value {
		x := v.Message().Interface().(*testpb.Message)
		x.Int32Field *= 2
		x.StringField = x.StringField + "/" + x.StringField
		return protoreflect.ValueOfMessage(x.ProtoReflect())
	})
	tt.runBasic(t)
	tt.runUpdate(t)
	tt.runMessage(t)
}

func TestBytesList(t *testing.T) {
	tt := newListTest("bytes", func(v protoreflect.Value) protoreflect.Value {
		x := v.Bytes()
		return protoreflect.ValueOfBytes(bytes.Join([][]byte{x, x}, []byte("/")))
	})
	tt.runBasic(t)
	tt.runUpdate(t)
}

type listTest struct {
	field    string
	fd       protoreflect.FieldDescriptor
	listOnly *testpb.Message
	mutate   func(protoreflect.Value) protoreflect.Value
}

func newListTest(valueType string, mutate func(protoreflect.Value) protoreflect.Value) *listTest {
	field := "repeated_" + valueType + "_field"
	fd := testMsg.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(field))
	listOnly := &testpb.Message{}
	dstList := listOnly.ProtoReflect().Mutable(fd).List()
	srcList := testMsg.ProtoReflect().Get(fd).List()
	for i, n := 0, srcList.Len(); i < n; i++ {
		dstList.Append(srcList.Get(i))
	}
	return &listTest{
		field:    field,
		fd:       fd,
		listOnly: listOnly,
		mutate:   mutate,
	}
}

func (tt *listTest) runBasic(t *testing.T) {
	basicTest{
		mask:  tt.field,
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.listOnly,
	}.run(t)

	basicTest{
		mask:  tt.field + "," + tt.field,
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.listOnly,
	}.run(t)

	basicTest{
		mask:  tt.field + ".*",
		paths: []string{tt.field},
		msg:   testMsg,
		out:   tt.listOnly,
	}.run(t)

	basicTest{
		mask: tt.field + ".invalid_subfield",
		err:  true,
	}.run(t)

	basicTest{
		mask: tt.field + ".*.invalid_subfield",
		err:  true,
	}.run(t)

	basicTest{
		mask: tt.field + "," + tt.field + ".invalid_subfield",
		err:  true,
	}.run(t)

	if tt.fd.Message() == nil {
		// TODO: Verify message subfields after they're wild-carded.
		basicTest{
			mask: tt.field + "," + tt.field + ".*.invalid_subfield",
			err:  true,
		}.run(t)
	}
}

func (tt *listTest) runUpdate(t *testing.T) {
	updateTest{
		name: tt.field + ":src-nil;replace",
		mask: tt.field,
		opts: []Option{WithUpdateRepeated(UpdateReplacesRepeated)},
		dst:  testMsg,
		src:  nil,
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.field + ":src-list-nil;append",
		mask: tt.field,
		opts: []Option{WithUpdateRepeated(UpdateAppendsRepeated)},
		dst:  testMsg,
		src:  nil,
		out:  testMsg,
	}.run(t)

	updateTest{
		name: tt.field + ":src-list-nil;replace",
		opts: []Option{WithUpdateRepeated(UpdateReplacesRepeated)},
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
		name: tt.field + ":src-list-nil;append",
		mask: tt.field,
		opts: []Option{WithUpdateRepeated(UpdateAppendsRepeated)},
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		out: testMsg,
	}.run(t)

	updateTest{
		name: tt.field + ":dst-nil;replace",
		mask: tt.field,
		opts: []Option{WithUpdateRepeated(UpdateReplacesRepeated)},
		dst: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		src: testMsg,
		out: testMsg,
	}.run(t)

	updateTest{
		name: tt.field + ":dst-list-nil;append",
		mask: tt.field,
		opts: []Option{WithUpdateRepeated(UpdateAppendsRepeated)},
		dst: func() *testpb.Message {
			out := clone(testMsg)
			out.ProtoReflect().Clear(tt.fd)
			return out
		}(),
		src: testMsg,
		out: testMsg,
	}.run(t)

	updateTest{
		name: tt.field + ":replace",
		mask: tt.field,
		opts: []Option{WithUpdateRepeated(UpdateReplacesRepeated)},
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			list := out.ProtoReflect().Get(tt.fd).List()
			for i, n := 0, list.Len(); i < n; i++ {
				list.Set(i, tt.mutate(list.Get(i)))
			}
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			list := out.ProtoReflect().Get(tt.fd).List()
			for i, n := 0, list.Len(); i < n; i++ {
				list.Set(i, tt.mutate(list.Get(i)))
			}
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.field + ":append",
		mask: tt.field,
		opts: []Option{WithUpdateRepeated(UpdateAppendsRepeated)},
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			list := out.ProtoReflect().Get(tt.fd).List()
			for i, n := 0, list.Len(); i < n; i++ {
				list.Set(i, tt.mutate(list.Get(i)))
			}
			return out
		}(),
		out: func() *testpb.Message {
			srcList := clone(testMsg).ProtoReflect().Mutable(tt.fd).List()
			out := clone(testMsg)
			outList := out.ProtoReflect().Mutable(tt.fd).List()
			for i, n := 0, srcList.Len(); i < n; i++ {
				outList.Append(tt.mutate(srcList.Get(i)))
			}
			return out
		}(),
	}.run(t)
}

func (tt *listTest) runMessage(t *testing.T) {
	basicTest{
		mask: tt.field + ".*.invalid_subfield",
		err:  true,
	}.run(t)

	basicTest{
		mask: tt.field + ".0.int32_field",
		err:  true,
	}.run(t)

	basicTest{
		mask:  tt.field + ".*.int32_field",
		paths: []string{tt.field + ".*.int32_field"},
		msg:   testMsg,
		out: func() *testpb.Message {
			out := clone(tt.listOnly)
			list := out.ProtoReflect().Get(tt.fd).List()
			for i, n := 0, list.Len(); i < n; i++ {
				src := list.Get(i).Message().Interface().(*testpb.Message)
				dst := &testpb.Message{
					Int32Field: src.Int32Field,
				}
				list.Set(i, protoreflect.ValueOfMessage(dst.ProtoReflect()))
			}
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.field + ".*.int32_field:replace",
		mask: tt.field + ".*.int32_field",
		opts: []Option{WithUpdateRepeated(UpdateReplacesRepeated)},
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(tt.listOnly)
			list := out.ProtoReflect().Get(tt.fd).List()
			for i, n := 0, list.Len(); i < n; i++ {
				m := list.Get(i).Message().Interface().(*testpb.Message)
				m.Int32Field *= 2
				m.StringField = "mutated-" + m.StringField
			}
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			list := out.ProtoReflect().Get(tt.fd).List()
			for i, n := 0, list.Len(); i < n; i++ {
				src := list.Get(i).Message().Interface().(*testpb.Message)
				msg := &testpb.Message{
					Int32Field: src.Int32Field * 2,
				}
				list.Set(i, protoreflect.ValueOfMessage(msg.ProtoReflect()))
			}
			return out
		}(),
	}.run(t)

	updateTest{
		name: tt.field + ".*.int32_field:append",
		mask: tt.field + ".*.int32_field",
		opts: []Option{WithUpdateRepeated(UpdateAppendsRepeated)},
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			list := out.ProtoReflect().Get(tt.fd).List()
			for i, n := 0, list.Len(); i < n; i++ {
				m := list.Get(i).Message().Interface().(*testpb.Message)
				m.Int32Field *= 2
				m.StringField = "mutated-" + m.StringField
			}
			return out
		}(),
		out: func() *testpb.Message {
			srcList := clone(testMsg).ProtoReflect().Mutable(tt.fd).List()
			out := clone(testMsg)
			outList := out.ProtoReflect().Mutable(tt.fd).List()
			for i, n := 0, srcList.Len(); i < n; i++ {
				src := srcList.Get(i).Message().Interface().(*testpb.Message)
				msg := &testpb.Message{
					Int32Field: src.Int32Field * 2,
				}
				outList.Append(protoreflect.ValueOfMessage(msg.ProtoReflect()))
			}
			out.ProtoReflect().Set(tt.fd, protoreflect.ValueOfList(outList))
			return out
		}(),
	}.run(t)
}
