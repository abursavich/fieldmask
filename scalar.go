// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ fieldMask = (*scalarFieldMask)(nil)

type scalarFieldMask struct {
	desc protoreflect.FieldDescriptor
}

func newScalarFieldMask(desc protoreflect.FieldDescriptor) *scalarFieldMask {
	return &scalarFieldMask{desc: desc}
}

func (fm *scalarFieldMask) complete() bool { return true }

func (fm *scalarFieldMask) init(path string) error { return addScalarPath(path) }

func (fm *scalarFieldMask) append(path string) error { return addScalarPath(path) }

func (fm *scalarFieldMask) paths() []string { return nil }

func (fm *scalarFieldMask) mask(protoreflect.Message, protoreflect.Value) { /* no-op */ }

func (fm *scalarFieldMask) update(parent protoreflect.Message, value protoreflect.Value, exists bool) {
	if !exists || !value.IsValid() {
		parent.Clear(fm.desc)
		return
	}
	parent.Set(fm.desc, value)
}

func (fm *scalarFieldMask) clone(parent protoreflect.Message, value protoreflect.Value) protoreflect.Value {
	if fm.desc.Kind() == protoreflect.BytesKind {
		return cloneBytesValue(value)
	}
	return value
}

func addScalarPath(path string) error {
	if path != "" {
		return fmt.Errorf("invalid scalar field subpath: %q", path)
	}
	return nil
}
