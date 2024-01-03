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

func newListFieldMask(settings *settings, desc protoreflect.FieldDescriptor) fieldMask {
	if isMessage(desc.Kind()) {
		// TODO: Test GroupKind.
		return &msgListFieldMask{desc: desc, settings: settings}
	}
	return &scalarListFieldMask{desc: desc, settings: settings}
}

var _ fieldMask = (*scalarListFieldMask)(nil)

type scalarListFieldMask struct {
	desc     protoreflect.FieldDescriptor
	settings *settings
}

func (fm *scalarListFieldMask) complete() bool { return true }

func (fm *scalarListFieldMask) init(path string) error {
	return fm.add(path)
}

func (fm *scalarListFieldMask) append(path string) error {
	return fm.add(path)
}

func (fm *scalarListFieldMask) add(path string) error {
	if path == "" || path == "*" {
		return nil
	}
	token, subpath, err := nextSegment(path)
	if err != nil {
		return err
	}
	if token != "*" {
		return fmt.Errorf("invalid list path: %q", path)
	}
	if subpath != "" {
		return fmt.Errorf("invalid scalar field subpath: %q", subpath)
	}
	return nil
}

func (fm *scalarListFieldMask) paths() []string {
	return nil
}

func (fm *scalarListFieldMask) mask(parent protoreflect.Message, value protoreflect.Value) {}

func (fm *scalarListFieldMask) clone(parent protoreflect.Message, value protoreflect.Value) protoreflect.Value {
	src := value.List()
	dst := parent.NewField(fm.desc).List()
	fm.settings.copyList(dst, src, fm.desc)
	return protoreflect.ValueOfList(dst)
}

func (fm *scalarListFieldMask) update(parent protoreflect.Message, value protoreflect.Value, exists bool) {
	if !exists || !value.IsValid() || !value.List().IsValid() {
		if fm.settings.updateRepeated == UpdateReplacesRepeated {
			parent.Clear(fm.desc)
		}
		return
	}

	switch fm.settings.updateRepeated {
	case UpdateAppendsRepeated:
		src := value.List()
		dst := parent.Mutable(fm.desc).List()
		for i, n := 0, src.Len(); i < n; i++ {
			dst.Append(src.Get(i))
		}
	default: // UpdateReplacesRepeated
		parent.Set(fm.desc, value)
	}
}

var _ fieldMask = (*msgListFieldMask)(nil)

type msgListFieldMask struct {
	desc     protoreflect.FieldDescriptor
	msgMask  *msgMask
	settings *settings
}

func (fm *msgListFieldMask) complete() bool { return fm.msgMask == nil }

func (fm *msgListFieldMask) init(path string) error {
	if path == "" || path == "*" {
		fm.msgMask = nil
		return nil
	}
	token, subpath, err := nextSegment(path)
	if err != nil {
		return err
	}
	if token != "*" {
		return fmt.Errorf("invalid list path: %q", path)
	}
	vm := newMsgMask(fm.settings, fm.desc.Message())
	if err := vm.init(subpath); err != nil {
		return err
	}
	fm.msgMask = vm
	return nil
}

func (fm *msgListFieldMask) append(path string) error {
	if path == "" || path == "*" {
		fm.msgMask = nil
		return nil
	}
	token, subpath, err := nextSegment(path)
	if err != nil {
		return err
	}
	if token != "*" {
		return fmt.Errorf("invalid list path: %q", path)
	}
	if fm.msgMask == nil {
		// TODO: Validate the subpath.
		return nil
	}
	return fm.msgMask.append(subpath)
}

func (fm *msgListFieldMask) paths() []string {
	if fm.complete() {
		return nil
	}
	subs := fm.msgMask.paths()
	paths := make([]string, len(subs))
	for i, sub := range subs {
		paths[i] = joinPath("*", sub)
	}
	return paths
}

func (fm *msgListFieldMask) mask(parent protoreflect.Message, value protoreflect.Value) {
	if fm.msgMask == nil {
		return
	}
	list := value.List()
	for i, n := 0, list.Len(); i < n; i++ {
		fm.msgMask.mask(list.Get(i).Message())
	}
}

func (fm *msgListFieldMask) clone(parent protoreflect.Message, value protoreflect.Value) protoreflect.Value {
	src := value.List()
	dst := parent.NewField(fm.desc).List()
	if fm.msgMask == nil {
		fm.settings.copyList(dst, src, fm.desc)
		return protoreflect.ValueOfList(dst)
	}
	for i, n := 0, src.Len(); i < n; i++ {
		msg := src.Get(i).Message()
		clone := fm.msgMask.clone(msg)
		dst.Append(protoreflect.ValueOfMessage(clone))
	}
	return protoreflect.ValueOfList(dst)
}

func (fm *msgListFieldMask) update(parent protoreflect.Message, value protoreflect.Value, exists bool) {
	if !exists || !value.IsValid() || !value.List().IsValid() {
		if fm.settings.updateRepeated == UpdateReplacesRepeated {
			parent.Clear(fm.desc)
		}
		return
	}

	if fm.complete() {
		fm.updateComplete(parent, value)
		return
	}
	src := value.List()
	dst := parent.Mutable(fm.desc).List()
	if fm.settings.updateRepeated == UpdateReplacesRepeated {
		dst.Truncate(0)
	}
	for i, n := 0, src.Len(); i < n; i++ {
		// TODO: This doesn't necessarily require a clone.
		msg := src.Get(i).Message()
		clone := fm.msgMask.clone(msg)
		dst.Append(protoreflect.ValueOfMessage(clone))
	}
}

func (fm *msgListFieldMask) updateComplete(parent protoreflect.Message, value protoreflect.Value) {
	switch fm.settings.updateRepeated {
	case UpdateAppendsRepeated:
		src := value.List()
		dst := parent.Mutable(fm.desc).List()
		for i, n := 0, src.Len(); i < n; i++ {
			dst.Append(src.Get(i))
		}
	default: // UpdateReplacesRepeated
		parent.Set(fm.desc, value)
	}
}

func isMessage(k protoreflect.Kind) bool {
	return k == protoreflect.MessageKind || k == protoreflect.GroupKind
}
