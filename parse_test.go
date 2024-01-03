// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"testing"
)

func TestNextPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		path string
		rest string
		err  error
	}{
		{
			name: "empty",
			err:  errSyntax,
		},
		{
			name: "dot",
			in:   ".",
			err:  errSyntax,
		},
		{
			name: "comma",
			in:   ",",
			err:  errSyntax,
		},
		{
			name: "asterisk",
			in:   "*",
			path: "*",
		},
		{
			name: "simple",
			in:   "foo",
			path: "foo",
		},
		{
			name: "trailing-dot",
			in:   "foo.",
			err:  errSyntax,
		},
		{
			name: "trailing-comma",
			in:   "foo,",
			err:  errSyntax,
		},
		{
			name: "multipart",
			in:   "foo.bar",
			path: "foo.bar",
		},
		{
			name: "quoted",
			in:   "`foo.bar`",
			path: "`foo.bar`",
		},
		{
			name: "unclosed-quote",
			in:   "`foo.bar",
			err:  errSyntax,
		},
		{
			name: "unseparated-quote",
			in:   "`foo.bar`123.qux",
			err:  errSyntax,
		},
		{
			name: "trailing-error",
			in:   "`foo.bar``123.qux",
			err:  errSyntax,
		},
		{
			name: "multipath",
			in:   "foo,bar",
			path: "foo",
			rest: "bar",
		},
		{
			name: "miltipart-multipath",
			in:   "foo.*.123,bar.qux",
			path: "foo.*.123",
			rest: "bar.qux",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, rest, err := nextPath(tt.in)
			if err != tt.err {
				t.Errorf("unexpected err: got: %v; want: %v", err, tt.err)
			}
			if path != tt.path {
				t.Errorf("unexpected path: got: %q; want: %q", path, tt.path)
			}
			if rest != tt.rest {
				t.Errorf("unexpected rest: got: %q; want: %q", rest, tt.rest)
			}
		})
	}
}

func TestNextSegment(t *testing.T) {
	tests := []struct {
		name string
		in   string
		part string
		rest string
		err  error
	}{
		{
			name: "empty",
			err:  errSyntax,
		},
		{
			name: "dot",
			in:   ".",
			err:  errSyntax,
		},
		{
			name: "comma",
			in:   ",",
			err:  errSyntax,
		},
		{
			name: "asterisk",
			in:   "*",
			part: "*",
		},
		{
			name: "simple",
			in:   "foo",
			part: "foo",
		},
		{
			name: "trailing-dot",
			in:   "foo.",
			err:  errSyntax,
		},
		{
			name: "trailing-comma",
			in:   "foo,",
			err:  errSyntax,
		},
		{
			name: "multipart",
			in:   "foo.bar.baz",
			part: "foo",
			rest: "bar.baz",
		},
		{
			name: "quoted",
			in:   "`foo.bar`",
			part: "`foo.bar`",
		},
		{
			name: "unclosed-quote",
			in:   "`foo.bar",
			err:  errSyntax,
		},
		{
			name: "unseparated-quote",
			in:   "`foo.bar`123.qux",
			err:  errSyntax,
		},
		{
			name: "trailing-error",
			in:   "`foo.bar``123.qux",
			err:  errSyntax,
		},
		{
			name: "multipath",
			in:   "foo,bar",
			err:  errSyntax,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part, rest, err := nextSegment(tt.in)
			if err != tt.err {
				t.Errorf("unexpected err: got: %v; want: %v", err, tt.err)
			}
			if part != tt.part {
				t.Errorf("unexpected part: got: %q; want: %q", part, tt.part)
			}
			if rest != tt.rest {
				t.Errorf("unexpected rest: got: %q; want: %q", rest, tt.rest)
			}
		})
	}
}
