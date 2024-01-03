// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func newMapFieldMask(settings *settings, desc protoreflect.FieldDescriptor) fieldMask {
	if isMessage(desc.MapValue().Kind()) {
		// TODO: Test GroupKind.
		return newMsgMapFieldMask(settings, desc)
	}
	return newScalarMapFieldMask(settings, desc)
}

type keyFuncs[T constraints.Ordered] struct {
	value  func(protoreflect.MapKey) T
	format func(T) string
	parse  func(string) (T, error)
}

func (fn *keyFuncs[T]) key(s string) (key T, err error) {
	if strings.HasPrefix(s, "`") {
		s, err = strconv.Unquote(s)
		if err != nil {
			return key, err
		}
	}
	return fn.parse(s)
}

var stringKeyFuncs = keyFuncs[string]{
	value:  protoreflect.MapKey.String,
	format: func(v string) string { return v },
	parse:  func(v string) (string, error) { return v, nil },
}

var int32KeyFuncs = keyFuncs[int32]{
	value:  func(v protoreflect.MapKey) int32 { return int32(v.Int()) },
	format: func(v int32) string { return strconv.FormatInt(int64(v), 10) },
	parse: func(s string) (int32, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	},
}

var int64KeyFuncs = keyFuncs[int64]{
	value:  protoreflect.MapKey.Int,
	format: func(v int64) string { return strconv.FormatInt(v, 10) },
	parse:  func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) },
}

var uint32KeyFuncs = keyFuncs[uint32]{
	value:  func(v protoreflect.MapKey) uint32 { return uint32(v.Uint()) },
	format: func(v uint32) string { return strconv.FormatUint(uint64(v), 10) },
	parse: func(s string) (uint32, error) {
		v, err := strconv.ParseUint(s, 10, 32)
		return uint32(v), err
	},
}

var uint64KeyFuncs = keyFuncs[uint64]{
	value:  protoreflect.MapKey.Uint,
	format: func(v uint64) string { return strconv.FormatUint(v, 10) },
	parse:  func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) },
}

var boolKeyFuncs = keyFuncs[byte]{
	value: func(v protoreflect.MapKey) byte {
		return boolToByte(v.Bool())
	},
	format: func(b byte) string {
		if b == 0 {
			return "false"
		}
		return "true"
	},
	parse: func(s string) (val byte, err error) {
		b, err := strconv.ParseBool(s)
		return boolToByte(b), err
	},
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

type scalarMapFieldMask[T constraints.Ordered] struct {
	desc protoreflect.FieldDescriptor
	keys map[T]bool
	keyFuncs[T]

	settings *settings
}

func newScalarMapFieldMask(settings *settings, desc protoreflect.FieldDescriptor) fieldMask {
	switch kind := desc.MapKey().Kind(); kind {
	case protoreflect.StringKind:
		return &scalarMapFieldMask[string]{
			desc:     desc,
			keyFuncs: stringKeyFuncs,
			settings: settings,
		}
	case protoreflect.BoolKind:
		// NOTE: We're using a byte because bool is not Ordered and we sort the keys when generating paths.
		return &scalarMapFieldMask[byte]{
			desc:     desc,
			keyFuncs: boolKeyFuncs,
			settings: settings,
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return &scalarMapFieldMask[int32]{
			desc:     desc,
			keyFuncs: int32KeyFuncs,
			settings: settings,
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return &scalarMapFieldMask[int64]{
			desc:     desc,
			keyFuncs: int64KeyFuncs,
			settings: settings,
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return &scalarMapFieldMask[uint32]{
			desc:     desc,
			keyFuncs: uint32KeyFuncs,
			settings: settings,
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return &scalarMapFieldMask[uint64]{
			desc:     desc,
			keyFuncs: uint64KeyFuncs,
			settings: settings,
		}
	default:
		panic(fmt.Sprintf("invalid map key kind: %v", kind))
	}
}

func (fm *scalarMapFieldMask[T]) complete() bool { return fm.keys == nil }

func (fm *scalarMapFieldMask[T]) init(path string) error {
	return fm.add(path)
}

func (fm *scalarMapFieldMask[T]) append(path string) error {
	if fm.complete() {
		return nil
	}
	return fm.add(path)
}

func (fm *scalarMapFieldMask[T]) add(path string) error {
	if path == "" || path == "*" {
		return fm.addWild("")
	}
	name, subpath, err := nextSegment(path)
	if err != nil {
		return err
	}
	if name == "*" {
		return fm.addWild(subpath)
	}
	return fm.addKeyed(name, subpath)
}

func (fm *scalarMapFieldMask[T]) addWild(subpath string) error {
	if subpath != "" {
		return fmt.Errorf("invalid scalar field subpath: %q", subpath)
	}
	fm.keys = nil
	return nil
}

func (fm *scalarMapFieldMask[T]) addKeyed(key, subpath string) error {
	k, err := fm.key(key)
	if err != nil {
		return err
	}
	if subpath != "" {
		return fmt.Errorf("invalid scalar field subpath: %q", subpath)
	}
	if fm.keys == nil {
		fm.keys = make(map[T]bool)
	}
	fm.keys[k] = true
	return nil
}

func (fm *scalarMapFieldMask[T]) paths() []string {
	if fm.keys == nil {
		return nil
	}
	keys := maps.Keys(fm.keys)
	slices.Sort(keys)
	paths := make([]string, len(keys))
	for i, key := range keys {
		paths[i] = maybeQuote(fm.format(key))
	}
	return paths
}

func (fm *scalarMapFieldMask[T]) mask(parent protoreflect.Message, value protoreflect.Value) {
	if fm.complete() {
		return
	}
	protoMap := value.Map()
	protoMap.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		if !fm.keys[fm.value(key)] {
			protoMap.Clear(key)
			return true
		}
		protoMap.Set(key, val)
		return true
	})
}

func (fm *scalarMapFieldMask[T]) clone(parent protoreflect.Message, value protoreflect.Value) protoreflect.Value {
	src := value.Map()
	dst := parent.NewField(fm.desc).Map()
	switch {
	case fm.complete():
		fm.settings.copyMap(dst, src, fm.desc)
	case fm.desc.MapValue().Kind() == protoreflect.BytesKind:
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			if fm.keys[fm.value(key)] {
				dst.Set(key, cloneBytesValue(val))
			}
			return true
		})
	default:
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			if fm.keys[fm.value(key)] {
				dst.Set(key, val)
			}
			return true
		})
	}
	return protoreflect.ValueOfMap(dst)
}

func (fm *scalarMapFieldMask[T]) update(parent protoreflect.Message, value protoreflect.Value, exists bool) {
	switch {
	case !value.IsValid() || !value.Map().IsValid():
		fm.clear(parent)
	case fm.complete():
		parent.Set(fm.desc, value)
	default:
		src := value.Map()
		dst := parent.Mutable(fm.desc).Map()
		dst.Range(func(key protoreflect.MapKey, _ protoreflect.Value) bool {
			// Remove values that have a mask but aren't in the src.
			if fm.keys[fm.value(key)] && !src.Has(key) {
				dst.Clear(key)
			}
			return true
		})
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			// Set values that have a mask.
			if fm.keys[fm.value(key)] {
				dst.Set(key, val)
			}
			return true
		})
	}
}

func (fm *scalarMapFieldMask[T]) clear(parent protoreflect.Message) {
	switch {
	case !parent.Has(fm.desc):
		// Nothing to clear
	case fm.complete():
		parent.Clear(fm.desc)
	default:
		dst := parent.Mutable(fm.desc).Map()
		dst.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			if _, ok := fm.keys[fm.value(key)]; ok { // if keyed mask DOES exists
				dst.Clear(key)
			}
			return true
		})
	}
}

type msgMapFieldMask[T constraints.Ordered] struct {
	desc       protoreflect.FieldDescriptor
	wildMask   *msgMask
	keyedMasks map[T]*msgMask
	keyFuncs[T]

	settings *settings
}

func newMsgMapFieldMask(settings *settings, desc protoreflect.FieldDescriptor) fieldMask {
	switch kind := desc.MapKey().Kind(); kind {
	case protoreflect.StringKind:
		return &msgMapFieldMask[string]{
			desc:     desc,
			keyFuncs: stringKeyFuncs,
			settings: settings,
		}
	case protoreflect.BoolKind:
		// NOTE: We're using a byte because bool is not Ordered and we sort the keys when generating paths.
		return &msgMapFieldMask[byte]{
			desc:     desc,
			keyFuncs: boolKeyFuncs,
			settings: settings,
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return &msgMapFieldMask[int32]{
			desc:     desc,
			keyFuncs: int32KeyFuncs,
			settings: settings,
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return &msgMapFieldMask[int64]{
			desc:     desc,
			keyFuncs: int64KeyFuncs,
			settings: settings,
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return &msgMapFieldMask[uint32]{
			desc:     desc,
			keyFuncs: uint32KeyFuncs,
			settings: settings,
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return &msgMapFieldMask[uint64]{
			desc:     desc,
			keyFuncs: uint64KeyFuncs,
			settings: settings,
		}
	default:
		panic(fmt.Sprintf("invalid map key kind: %v", kind))
	}
}

func (fm *msgMapFieldMask[T]) complete() bool { return fm.keyedMasks == nil && fm.wildMask == nil }

func (fm *msgMapFieldMask[T]) init(path string) error {
	return fm.add(path)
}

func (fm *msgMapFieldMask[T]) append(path string) error {
	if fm.complete() {
		return nil
	}
	return fm.add(path)
}

func (fm *msgMapFieldMask[T]) add(path string) error {
	if path == "" || path == "*" {
		return fm.addWild("")
	}
	name, subpath, err := nextSegment(path)
	if err != nil {
		return err
	}
	if name == "*" {
		return fm.addWild(subpath)
	}
	return fm.addKeyed(name, subpath)
}

func (fm *msgMapFieldMask[T]) addWild(subpath string) error {
	if subpath == "" {
		fm.wildMask = nil
		fm.keyedMasks = nil
		return nil
	}
	if fm.wildMask == nil {
		m := newMsgMask(fm.settings, fm.desc.MapValue().Message())
		if err := m.init(subpath); err != nil {
			return err
		}
		fm.wildMask = m
	} else if err := fm.wildMask.append(subpath); err != nil {
		return err
	}
	for _, m := range fm.keyedMasks {
		if err := m.append(subpath); err != nil {
			panic(fmt.Sprintf("fieldmask: internal error: successful wild mask append failed on keyed mask: %q: %v", subpath, err))
		}
	}
	return nil
}

func (fm *msgMapFieldMask[T]) addKeyed(key, subpath string) error {
	k, err := fm.key(key)
	if err != nil {
		return err
	}
	if m, ok := fm.keyedMasks[k]; ok {
		return m.append(subpath)
	}

	m := newMsgMask(fm.settings, fm.desc.MapValue().Message())
	if err := m.init(subpath); err != nil {
		return err
	}
	if fm.wildMask != nil {
		for _, path := range fm.wildMask.paths() {
			if err := m.append(path); err != nil {
				panic(fmt.Sprintf("fieldmask: internal error: successful wild mask append failed on keyed mask: %q: %v", subpath, err))
			}
		}
	}
	if fm.keyedMasks == nil {
		fm.keyedMasks = make(map[T]*msgMask)
	}
	fm.keyedMasks[k] = m
	return nil
}

func (fm *msgMapFieldMask[T]) paths() []string {
	var wild []string
	var paths []string
	if fm.wildMask != nil {
		wild = fm.wildMask.paths()
		for _, sub := range wild {
			paths = append(paths, joinPath("*", sub))
		}
	}
	if fm.keyedMasks == nil {
		return paths
	}
	var needles map[string]bool
	lazyNeedles := false
	keys := maps.Keys(fm.keyedMasks)
	slices.Sort(keys)
	for _, key := range keys {
		name := maybeQuote(fm.format(key))
		subs := fm.keyedMasks[key].paths()
		if len(subs) == 0 {
			paths = append(paths, name)
			continue
		}
		if !lazyNeedles {
			needles = toSet(wild)
			lazyNeedles = true
		}
		for _, sub := range remove(subs, needles) {
			paths = append(paths, joinPath(name, sub))
		}
	}
	return paths
}

func (fm *msgMapFieldMask[T]) lookupMask(key protoreflect.MapKey) (*msgMask, bool) {
	if fm.keyedMasks != nil {
		if m, ok := fm.keyedMasks[fm.value(key)]; ok {
			return m, true
		}
	}
	if fm.wildMask != nil {
		return fm.wildMask, true
	}
	return nil, false
}

func (fm *msgMapFieldMask[T]) mask(parent protoreflect.Message, value protoreflect.Value) {
	if fm.complete() {
		return
	}
	protoMap := value.Map()
	protoMap.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		m, ok := fm.lookupMask(key)
		if !ok {
			protoMap.Clear(key)
			return true
		}
		m.mask(val.Message())
		return true
	})
}

func (fm *msgMapFieldMask[T]) clone(parent protoreflect.Message, value protoreflect.Value) protoreflect.Value {
	src := value.Map()
	dst := parent.NewField(fm.desc).Map()
	switch {
	case fm.complete():
		fm.settings.copyMap(dst, src, fm.desc)
	default:
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			if m, ok := fm.lookupMask(key); ok {
				dst.Set(key, protoreflect.ValueOfMessage(m.clone(val.Message())))
			}
			return true
		})
	}
	return protoreflect.ValueOfMap(dst)
}

func (fm *msgMapFieldMask[T]) update(parent protoreflect.Message, value protoreflect.Value, exists bool) {
	switch {
	case !value.IsValid() || !value.Map().IsValid():
		fm.clear(parent)
	case fm.complete():
		fm.settings.updateMap(parent.Mutable(fm.desc).Map(), value.Map(), fm.desc)
	default:
		src := value.Map()
		dst := parent.Mutable(fm.desc).Map()
		dst.Range(func(key protoreflect.MapKey, _ protoreflect.Value) bool {
			// Remove values that have a mask but aren't in the src.
			if _, ok := fm.lookupMask(key); ok && !src.Has(key) {
				dst.Clear(key)
			}
			return true
		})
		src.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			// Update values that have a mask.
			if m, ok := fm.lookupMask(key); ok {
				m.update(dst.Mutable(key).Message(), val.Message())
			}
			return true
		})
	}
}

func (fm *msgMapFieldMask[T]) clear(parent protoreflect.Message) {
	switch {
	case !parent.Has(fm.desc):
		// Nothing to clear
	case fm.complete() || fm.wildMask != nil:
		parent.Clear(fm.desc)
	default:
		dst := parent.Mutable(fm.desc).Map()
		dst.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
			if _, ok := fm.keyedMasks[fm.value(key)]; ok { // if keyed mask DOES exists
				dst.Clear(key)
			}
			return true
		})
	}
}

func remove(haystack []string, needles map[string]bool) []string {
	if len(needles) == 0 {
		return haystack

	}
	return slices.DeleteFunc(haystack, func(v string) bool {
		return needles[v]
	})
}

func toSet[T comparable](s []T) map[T]bool {
	if len(s) == 0 {
		return nil
	}
	m := make(map[T]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}
