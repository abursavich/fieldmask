// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"bursavich.dev/fieldmask/internal/quote"
)

var errSyntax = strconv.ErrSyntax

func nextPath(s string) (path, rest string, err error) {
	if s == "" {
		return "", "", errSyntax
	}
	rest = s
	for {
		var tok string

		tok, rest, err = nextToken(rest)
		if err != nil || tok == "." || tok == "," {
			return "", "", errSyntax
		}
		if rest == "" {
			return s, "", nil
		}

		tok, rest, err = nextToken(rest)
		if err != nil || rest == "" {
			return "", "", errSyntax
		}
		if tok == "," {
			return s[:len(s)-len(rest)-1], rest, nil
		}
		if tok != "." {
			return "", "", errSyntax
		}
	}
}

func nextSegment(s string) (segment, rest string, err error) {
	segment, rest, err = nextToken(s)
	if err != nil || segment == "." || segment == "," {
		return "", "", errSyntax
	}
	if rest == "" {
		return segment, "", nil
	}
	next, rest, err := nextToken(rest)
	if err != nil || next != "." || rest == "" {
		return "", "", errSyntax
	}
	return segment, rest, nil
}

func nextToken(s string) (token, rest string, err error) {
	if s == "" {
		return "", "", errSyntax
	}
	switch s[0] {
	case '.', ',', '*':
		return s[0:1], s[1:], nil
	case '`':
		quoted, err := strconv.QuotedPrefix(s)
		if err != nil {
			return "", "", errSyntax
		}
		return quoted, s[len(quoted):], nil
	default:
		if i := strings.IndexAny(s, ".,"); i != -1 {
			return s[:i], s[i:], nil
		}
		return s, "", nil
	}
}

func joinPath(a, b string) string {
	return a + "." + b
}

func maybeQuote(segment string) string {
	if shouldQuote(segment) {
		return quote.With(segment, '`')
	}
	return segment
}

func shouldQuote(s string) bool {
	if s == "" || s == "*" {
		return true
	}
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
			if width == 1 && r == utf8.RuneError {
				return true
			}
		}
		if r == '.' || r == ',' || r == '`' {
			return true
		}
		if unicode.IsControl(r) || !strconv.IsPrint(r) {
			return true
		}
	}
	return false
}
