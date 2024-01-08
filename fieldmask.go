// SPDX-License-Identifier: MIT
//
// Copyright 2024 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package fieldmask

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// https://google.aip.dev/161 Field masks
// https://google.aip.dev/203 Field behavior documentation (TODO: IMMUTABLE, OUTPUT_ONLY. INPUT_ONLY)
// https://google.aip.dev/134 Standard methods: Update

type Option interface{ applyOption(*settings) }

type optionFunc func(*settings)

func (fn optionFunc) applyOption(s *settings) { fn(s) }

// WithExtensions returns an option that sets whether extensions are allowed.
func WithExtensions(allow bool) Option {
	return optionFunc(func(s *settings) { s.extensions = allow })
}

// FieldName specifies which field name to prefer when parsing and outputting paths.
type FieldName int

const (
	// TextFieldName uses the text field name, which is typically lower_snake_case.
	// This is the default behavior.
	TextFieldName FieldName = iota
	// JSONFieldName uses the JSON field name, which is typically lowerCamelCase.
	JSONFieldName
)

// WithFieldName returns an option that sets the given mode for field names.
// Either name is accepted when parsing paths, but only one is used for outputing paths.
func WithFieldName(mode FieldName) Option {
	return optionFunc(func(s *settings) {
		switch mode {
		case TextFieldName:
			s.lookupField = lookupTextField
		case JSONFieldName:
			s.lookupField = lookupJSONField
		}
	})
}

// MaskUnknowns specifies how to handle unknown fields when a message is masked.
type MaskUnknowns int

const (
	// MaskRemovesUnknowns removes any unknown fields when a message is masked.
	// This is the default behavior.
	MaskRemovesUnknowns MaskUnknowns = iota
	// MaskRetainsUnknowns retains any unknown fields when a message is masked.
	MaskRetainsUnknowns
)

// WithMaskUnknowns returns an option that sets the given mode for masking unknown fields.
func WithMaskUnknowns(mode MaskUnknowns) Option {
	return optionFunc(func(s *settings) { s.maskUnknowns = mode })
}

// UpdateUnknowns specifies how to update unknown fields.
type UpdateUnknowns int

const (
	// UpdateRetainsUnknowns retains any unknown fields on the destination message
	// when it's updated. This is the default behavior.
	UpdateRetainsUnknowns UpdateUnknowns = iota
	// UpdateAppendsUnknowns appends any unknown fields from the source message
	// to any unknown fields on the destination message when it's updated.
	UpdateAppendsUnknowns
	// UpdateReplacesUnknowns replaces any unknown fields on the destination message
	// with any unknown fields from the source message when it's updated.
	UpdateReplacesUnknowns
)

// WithUpdateUnknowns returns an option that sets the given mode for updating unknown fields.
func WithUpdateUnknowns(mode UpdateUnknowns) Option {
	return optionFunc(func(s *settings) { s.updateUnknowns = mode })
}

// UpdateRepeated specifies how to update repeated fields.
type UpdateRepeated int

const (
	// UpdateReplacesRepeated replaces any repeated fields on an update.
	// This is the default behavior.
	UpdateReplacesRepeated UpdateRepeated = iota
	// UpdateAppendsRepeated appends any repeated fields on an update.
	UpdateAppendsRepeated
)

// WithUpdateRepeated returns an option that sets the given mode for updating repeated fields.
func WithUpdateRepeated(mode UpdateRepeated) Option {
	return optionFunc(func(s *settings) { s.updateRepeated = mode })
}

type FieldMask[T proto.Message] struct {
	settings
	msg *msgMask
}

func newFieldMaskT[T proto.Message](options []Option) *FieldMask[T] {
	fm := FieldMask[T]{
		settings: settings{
			lookupField: lookupTextField,
		},
	}
	for _, o := range options {
		o.applyOption(&fm.settings)
	}
	var zero T
	fm.msg = newMsgMask(&fm.settings, zero.ProtoReflect().Descriptor())
	return &fm
}

func New[T proto.Message](paths []string, options ...Option) (*FieldMask[T], error) {
	fm := newFieldMaskT[T](options)
	if len(paths) == 0 {
		return fm, nil
	}
	if err := fm.msg.init(paths[0]); err != nil {
		return nil, err
	}
	for _, path := range paths[1:] {
		if err := fm.msg.append(path); err != nil {
			return nil, err
		}
	}
	return fm, nil
}

func FromProto[T proto.Message](fieldMask *fieldmaskpb.FieldMask, options ...Option) (*FieldMask[T], error) {
	return New[T](fieldMask.GetPaths(), options...)
}

func Parse[T proto.Message](paths string, options ...Option) (*FieldMask[T], error) {
	fm := newFieldMaskT[T](options)
	apply := fm.msg.init
	for {
		path, rest, err := nextPath(paths)
		if err != nil {
			return nil, err
		}
		if err := apply(path); err != nil {
			return nil, err
		}
		if rest == "" {
			return fm, nil
		}
		paths = rest
		apply = fm.msg.append
	}
}

func (fm *FieldMask[T]) Append(path string) error {
	return fm.msg.append(path)
}

func (fm *FieldMask[T]) Paths() []string {
	if paths := fm.msg.paths(); len(paths) > 0 {
		return paths
	}
	return []string{"*"}
}

func (fm *FieldMask[T]) Proto() *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{
		Paths: fm.Paths(),
	}
}

func (fm *FieldMask[T]) String() string {
	return strings.Join(fm.Paths(), ",")
}

func (fm *FieldMask[T]) Mask(msg T) {
	fm.msg.mask(msg.ProtoReflect())
}

func (fm *FieldMask[T]) Clone(msg T) T {
	return fm.msg.clone(msg.ProtoReflect()).Interface().(T)
}

func (fm *FieldMask[T]) Update(dst, src T) error {
	fm.msg.update(dst.ProtoReflect(), src.ProtoReflect())
	return nil
}

type fieldMask interface {
	// complete returns a value indicating if the full value is retained.
	complete() bool
	// init adds the first path to the mask.
	init(path string) error
	// append adds an additional path to the mask.
	append(path string) error
	// paths returns the simplified paths of the mask.
	paths() []string

	// mask masks the value in place.
	mask(parent protoreflect.Message, value protoreflect.Value)
	// update updates the parent with the masked version of the value.
	update(parent protoreflect.Message, value protoreflect.Value, exists bool)
	// clone returns a cloned and masked version of the value.
	clone(parent protoreflect.Message, value protoreflect.Value) protoreflect.Value
}

func newFieldMask(settings *settings, desc protoreflect.FieldDescriptor) fieldMask {
	if desc.IsList() {
		return newListFieldMask(settings, desc)
	}
	if desc.IsMap() {
		return newMapFieldMask(settings, desc)
	}
	if desc.Message() != nil {
		return newMsgFieldMask(settings, desc)
	}
	return newScalarFieldMask(desc)
}
