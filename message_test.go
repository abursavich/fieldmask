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

func TestMessage(t *testing.T) {
	basicTest{
		mask:  "*",
		paths: []string{"*"},
		msg:   testMsg,
		out:   testMsg,
	}.run(t)

	basicTest{
		mask:  "*,message_field",
		paths: []string{"*"},
		msg:   testMsg,
		out:   testMsg,
	}.run(t)

	basicTest{
		mask:  "message_field,*",
		paths: []string{"*"},
		msg:   testMsg,
		out:   testMsg,
	}.run(t)

	basicTest{
		mask: "invalid_init_field",
		err:  true,
	}.run(t)

	basicTest{
		name: "invalidInitField",
		mask: "invalidInitField",
		err:  true,
		opts: []Option{WithFieldName(JSONFieldName)},
	}.run(t)

	basicTest{
		mask: "message_field,invalid_append_field",
		err:  true,
	}.run(t)

	basicTest{
		name: "messageField,invalidInitField",
		mask: "messageField,invalidInitField",
		err:  true,
		opts: []Option{WithFieldName(JSONFieldName)},
	}.run(t)

	// TODO: Validate subpaths after wildcards.
	// basicTest{
	// 	mask: "*,invalid_field",
	// 	err:  true,
	// }.run(t)

	basicTest{
		mask:  "message_field",
		paths: []string{"message_field"},
		msg:   testMsg,
		out: &testpb.Message{
			MessageField: testMsg.MessageField,
		},
	}.run(t)

	updateTest{
		name: "message_field:nil-src",
		mask: "message_field",
		dst:  testMsg,
		src:  nil,
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.MessageField = nil
			return out
		}(),
	}.run(t)

	updateTest{
		name: "message_field:nil-src-field",
		mask: "message_field",
		dst:  testMsg,
		src: func() *testpb.Message {
			out := clone(testMsg)
			out.MessageField = nil
			return out
		}(),
		out: func() *testpb.Message {
			out := clone(testMsg)
			out.MessageField = nil
			return out
		}(),
	}.run(t)

	updateTest{
		name: "message_field:nil-dst-field",
		mask: "message_field",
		dst: func() *testpb.Message {
			out := clone(testMsg)
			out.MessageField = nil
			return out
		}(),
		src: testMsg,
		out: testMsg,
	}.run(t)
}
