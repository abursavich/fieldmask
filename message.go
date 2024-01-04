// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ fieldMask = (*msgFieldMask)(nil)

type msgFieldMask struct {
	desc protoreflect.FieldDescriptor
	msgMask
}

func newMsgFieldMask(settings *settings, desc protoreflect.FieldDescriptor) *msgFieldMask {
	return &msgFieldMask{
		desc: desc,
		msgMask: msgMask{
			desc:     desc.Message(),
			settings: settings,
		},
	}
}

func (fm *msgFieldMask) mask(parent protoreflect.Message, value protoreflect.Value) {
	fm.msgMask.mask(value.Message())
}

func (fm *msgFieldMask) clone(parent protoreflect.Message, value protoreflect.Value) protoreflect.Value {
	return protoreflect.ValueOfMessage(fm.msgMask.clone(value.Message()))
}

func (fm *msgFieldMask) update(parent protoreflect.Message, value protoreflect.Value, exists bool) {
	if !exists || !value.IsValid() {
		parent.Clear(fm.desc)
		return
	}
	src := value.Message()
	dst := parent.Mutable(fm.desc).Message()
	fm.msgMask.update(dst, src)
}

type msgMask struct {
	desc     protoreflect.MessageDescriptor
	fldDescs protoreflect.FieldDescriptors
	fields   map[string]fieldMask
	settings *settings
}

func newMsgMask(settings *settings, desc protoreflect.MessageDescriptor) *msgMask {
	return &msgMask{
		desc:     desc,
		fldDescs: desc.Fields(),
		settings: settings,
	}
}

func (mm *msgMask) complete() bool { return mm.fields == nil }

func (mm *msgMask) init(path string) error {
	if path == "" || path == "*" {
		return nil
	}
	name, subpath, err := nextSegment(path)
	if err != nil {
		return err
	}
	key, fd, ok := mm.settings.lookupField(mm.fldDescs, name)
	if !ok {
		return fmt.Errorf("unknown %v field: %q", mm.desc.FullName(), name)
	}
	fld := newFieldMask(mm.settings, fd)
	if err := fld.init(subpath); err != nil {
		return err
	}
	mm.fields = map[string]fieldMask{
		key: fld,
	}
	return nil
}

func (mm *msgMask) append(path string) error {
	if path == "" || path == "*" {
		mm.fields = nil
		return nil
	}
	name, subpath, err := nextSegment(path)
	if err != nil {
		return err
	}
	key, fd, ok := mm.settings.lookupField(mm.fldDescs, name)
	if !ok {
		return fmt.Errorf("unknown %v field: %q", mm.desc.FullName(), name)
	}
	if mm.fields == nil {
		// TODO: Validate the subpath.
		return nil
	}
	if fld, ok := mm.fields[key]; ok {
		return fld.append(subpath)
	}
	fld := newFieldMask(mm.settings, fd)
	if err := fld.init(subpath); err != nil {
		return err
	}
	mm.fields[key] = fld
	return nil
}

func (mm *msgMask) paths() []string {
	var paths []string
	names := maps.Keys(mm.fields)
	sort.Strings(names)
	for _, names := range names {
		subs := mm.fields[names].paths()
		for _, sub := range subs {
			paths = append(paths, joinPath(names, sub))
		}
		if len(subs) == 0 {
			paths = append(paths, names)
		}
	}
	return paths
}

func (mm *msgMask) mask(msg protoreflect.Message) {
	if mm.complete() {
		return
	}
	msg.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		if f, ok := mm.fields[string(fd.Name())]; ok && mm.settings.allow(fd) {
			f.mask(msg, val)
			return true
		}
		msg.Clear(fd)
		return true
	})
	if mm.settings.maskUnknowns != MaskRetainsUnknowns {
		msg.SetUnknown(nil)
	}
}

func (mm *msgMask) clone(msg protoreflect.Message) protoreflect.Message {
	out := msg.New()
	if mm.complete() {
		mm.settings.copyMessage(out, msg)
		return out
	}
	msg.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		if f, ok := mm.fields[string(fd.Name())]; ok && mm.settings.allow(fd) {
			out.Set(fd, f.clone(msg, val))
		}
		return true
	})
	if mm.settings.maskUnknowns == MaskRetainsUnknowns {
		out.SetUnknown(copyBytes(msg.GetUnknown()))
	}
	return out
}

func (mm *msgMask) update(dst, src protoreflect.Message) {
	if mm.complete() {
		mm.settings.updateMessage(dst, src)
		return
	}
	for name, mask := range mm.fields {
		_, fd, _ := mm.settings.lookupField(mm.fldDescs, name)
		mask.update(dst, src.Get(fd), src.Has(fd))
	}
	mm.settings.doUpdateUnknowns(dst, src)
}
