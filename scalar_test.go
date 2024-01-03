// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"testing"

	"bursavich.dev/fieldmask/internal/testpb"
)

func TestBool(t *testing.T) {
	basicTest{
		mask:  "bool_field",
		paths: []string{"bool_field"},
		msg:   testMsg,
		out: &testpb.Message{
			BoolField: testMsg.BoolField,
		},
	}.run(t)

	basicTest{
		mask: "bool_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestString(t *testing.T) {
	basicTest{
		mask:  "string_field",
		paths: []string{"string_field"},
		msg:   testMsg,
		out: &testpb.Message{
			StringField: testMsg.StringField,
		},
	}.run(t)

	basicTest{
		mask: "string_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestInt32(t *testing.T) {
	basicTest{
		mask:  "int32_field",
		paths: []string{"int32_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Int32Field: testMsg.Int32Field,
		},
	}.run(t)

	basicTest{
		mask: "int32_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestInt64(t *testing.T) {
	basicTest{
		mask:  "int64_field",
		paths: []string{"int64_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Int64Field: testMsg.Int64Field,
		},
	}.run(t)

	basicTest{
		mask: "int64_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestSint32(t *testing.T) {
	basicTest{
		mask:  "sint32_field",
		paths: []string{"sint32_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Sint32Field: testMsg.Sint32Field,
		},
	}.run(t)

	basicTest{
		mask: "sint32_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestSint64(t *testing.T) {
	basicTest{
		mask:  "sint64_field",
		paths: []string{"sint64_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Sint64Field: testMsg.Sint64Field,
		},
	}.run(t)

	basicTest{
		mask: "sint64_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestUint32(t *testing.T) {
	basicTest{
		mask:  "uint32_field",
		paths: []string{"uint32_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Uint32Field: testMsg.Uint32Field,
		},
	}.run(t)

	basicTest{
		mask: "uint32_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestUint64(t *testing.T) {
	basicTest{
		mask:  "uint64_field",
		paths: []string{"uint64_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Uint64Field: testMsg.Uint64Field,
		},
	}.run(t)

	basicTest{
		mask: "uint64_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestFixed32(t *testing.T) {
	basicTest{
		mask:  "fixed32_field",
		paths: []string{"fixed32_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Fixed32Field: testMsg.Fixed32Field,
		},
	}.run(t)

	basicTest{
		mask: "fixed32_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestFixed64(t *testing.T) {
	basicTest{
		mask:  "fixed64_field",
		paths: []string{"fixed64_field"},
		msg:   testMsg,
		out: &testpb.Message{
			Fixed64Field: testMsg.Fixed64Field,
		},
	}.run(t)

	basicTest{
		mask: "fixed64_field.invalid_subfield",
		err:  true,
	}.run(t)
}

func TestBytes(t *testing.T) {
	basicTest{
		mask:  "bytes_field",
		paths: []string{"bytes_field"},
		msg:   testMsg,
		out: &testpb.Message{
			BytesField: testMsg.BytesField,
		},
	}.run(t)

	basicTest{
		mask: "bytes_field.invalid_subfield",
		err:  true,
	}.run(t)
}
