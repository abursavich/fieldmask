// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldLookupFunc func(fields protoreflect.FieldDescriptors, name string) (key string, fd protoreflect.FieldDescriptor, found bool)

func lookupTextField(fields protoreflect.FieldDescriptors, name string) (key string, fd protoreflect.FieldDescriptor, found bool) {
	fd = fields.ByTextName(name)
	if fd == nil {
		fd = fields.ByJSONName(name)
	}
	if fd == nil {
		return "", nil, false
	}
	return fd.TextName(), fd, true
}

func lookupTextFieldStrict(fields protoreflect.FieldDescriptors, name string) (key string, fd protoreflect.FieldDescriptor, found bool) {
	fd = fields.ByTextName(name)
	if fd == nil || fd.TextName() != name {
		return "", nil, false
	}
	return fd.TextName(), fd, true
}

func lookupJSONField(fields protoreflect.FieldDescriptors, name string) (key string, fd protoreflect.FieldDescriptor, found bool) {
	fd = fields.ByJSONName(name)
	if fd == nil {
		fd = fields.ByTextName(name)
	}
	if fd == nil {
		return "", nil, false
	}
	return fd.JSONName(), fd, true
}

func lookupJSONFieldStrict(fields protoreflect.FieldDescriptors, name string) (key string, fd protoreflect.FieldDescriptor, found bool) {
	fd = fields.ByJSONName(name)
	if fd == nil || fd.JSONName() != name {
		return "", nil, false
	}
	return fd.JSONName(), fd, true
}

type settings struct {
	rootDesc   protoreflect.MessageDescriptor
	extensions bool

	lookupField    fieldLookupFunc
	maskUnknowns   MaskUnknowns
	updateUnknowns UpdateUnknowns
	updateRepeated UpdateRepeated
}

func (s *settings) allow(fd protoreflect.FieldDescriptor) bool {
	return !(fd.IsExtension() && !s.extensions)
}

func (s *settings) copyMessage(dst, src protoreflect.Message) {
	src.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		switch {
		case !s.allow(fd):
			// no-op
		case fd.IsList():
			s.copyList(dst.Mutable(fd).List(), val.List(), fd)
		case fd.IsMap():
			s.copyMap(dst.Mutable(fd).Map(), val.Map(), fd)
		case fd.Message() != nil:
			s.copyMessage(dst.Mutable(fd).Message(), val.Message())
		case fd.Kind() == protoreflect.BytesKind:
			dst.Set(fd, cloneBytesValue(val))
		default:
			dst.Set(fd, val)
		}
		return true
	})
	if s.maskUnknowns == MaskRetainsUnknowns {
		dst.SetUnknown(copyBytes(src.GetUnknown()))
	}
}

func (s *settings) copyList(dst, src protoreflect.List, fd protoreflect.FieldDescriptor) {
	switch {
	case fd.Message() != nil:
		for i, n := 0, src.Len(); i < n; i++ {
			msg := dst.NewElement()
			s.copyMessage(msg.Message(), src.Get(i).Message())
			dst.Append(msg)
		}
	case fd.Kind() == protoreflect.BytesKind:
		for i, n := 0, src.Len(); i < n; i++ {
			dst.Append(cloneBytesValue(src.Get(i)))
		}
	default:
		for i, n := 0, src.Len(); i < n; i++ {
			dst.Append(src.Get(i))
		}
	}
}

func (s *settings) copyMap(dst, src protoreflect.Map, fd protoreflect.FieldDescriptor) {
	fd = fd.MapValue()
	switch {
	case fd.Message() != nil:
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			msg := dst.NewValue()
			s.copyMessage(msg.Message(), val.Message())
			dst.Set(key, msg)
			return true
		})
	case fd.Kind() == protoreflect.BytesKind:
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			dst.Set(key, cloneBytesValue(val))
			return true
		})
	default:
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			dst.Set(key, val)
			return true
		})
	}
}

func (s *settings) updateMessage(dst, src protoreflect.Message) {
	fds := dst.Descriptor().Fields()
	for i, n := 0, fds.Len(); i < n; i++ {
		s.updateField(dst, src, fds.Get(i))
	}
	s.doUpdateUnknowns(dst, src)
}

func (s *settings) doUpdateUnknowns(dst, src protoreflect.Message) {
	var srcUnknowns protoreflect.RawFields
	if src.IsValid() {
		srcUnknowns = src.GetUnknown()
	}
	switch {
	case s.updateUnknowns == UpdateAppendsUnknowns && len(srcUnknowns) > 0:
		dst.SetUnknown(append(copyBytes(dst.GetUnknown()), srcUnknowns...))
	case s.updateUnknowns == UpdateReplacesUnknowns:
		dst.SetUnknown(copyBytes(srcUnknowns))
	}
}

func (s *settings) updateField(dst, src protoreflect.Message, fd protoreflect.FieldDescriptor) {
	if !s.allow(fd) {
		return // no-op
	}
	if !src.Has(fd) {
		if fd.IsList() && s.updateRepeated == UpdateAppendsRepeated {
			return // no-op
		}
		dst.Clear(fd)
		return
	}
	switch {
	case fd.IsList():
		s.updateList(dst.Mutable(fd).List(), src.Get(fd).List(), fd)
	case fd.IsMap():
		s.updateMap(dst.Mutable(fd).Map(), src.Get(fd).Map(), fd)
	case fd.Message() != nil:
		s.updateMessage(dst.Mutable(fd).Message(), src.Get(fd).Message())
	default:
		if src.Has(fd) {
			dst.Set(fd, src.Get(fd))
		} else {
			dst.Clear(fd)
		}
	}
}

func (s *settings) updateList(dst, src protoreflect.List, fd protoreflect.FieldDescriptor) {
	if s.updateRepeated != UpdateAppendsRepeated {
		dst.Truncate(0)
	}
	if fd.Message() != nil {
		for i, n := 0, src.Len(); i < n; i++ {
			// TODO: This doesn't necessarily require a copy.
			msg := dst.NewElement()
			s.updateMessage(msg.Message(), src.Get(i).Message())
			dst.Append(msg)
		}
		return
	}
	for i, n := 0, src.Len(); i < n; i++ {
		dst.Append(src.Get(i))
	}
}

func (s *settings) updateMap(dst, src protoreflect.Map, fd protoreflect.FieldDescriptor) {
	dst.Range(func(key protoreflect.MapKey, _ protoreflect.Value) bool {
		if !src.Has(key) {
			dst.Clear(key)
		}
		return true
	})
	if fd.MapValue().Message() != nil {
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			// TODO: This doesn't necessarily require a copy.
			msg := dst.NewValue()
			s.updateMessage(msg.Message(), val.Message())
			dst.Set(key, msg)
			return true
		})
		return
	}
	src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		dst.Set(key, val)
		return true
	})
}

func cloneBytesValue(val protoreflect.Value) protoreflect.Value {
	return protoreflect.ValueOfBytes(copyBytes(val.Bytes()))
}

func copyBytes(buf []byte) []byte {
	if buf == nil {
		return nil
	}
	dst := make([]byte, len(buf))
	copy(dst, buf)
	return dst
}
