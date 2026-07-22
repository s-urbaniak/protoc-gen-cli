package irbuild

import (
	"fmt"
	"slices"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Flag names the generated code already reserve.
var reservedFlagNames = []string{"filename", "input", "help", "output"}

// messageBinds is protojson's message types rendered as a single JSON value.
var messageBinds = map[protoreflect.FullName]ir.Bind{
	"google.protobuf.Timestamp": ir.BindString,
	"google.protobuf.Duration":  ir.BindString,
	"google.protobuf.FieldMask": ir.BindString,

	"google.protobuf.BoolValue":   ir.BindBool,
	"google.protobuf.StringValue": ir.BindString,
	"google.protobuf.BytesValue":  ir.BindString,
	"google.protobuf.Int32Value":  ir.BindInt,
	"google.protobuf.Int64Value":  ir.BindInt,
	"google.protobuf.UInt32Value": ir.BindUint,
	"google.protobuf.UInt64Value": ir.BindUint,
	"google.protobuf.FloatValue":  ir.BindFloat,
	"google.protobuf.DoubleValue": ir.BindFloat,

	"google.protobuf.Struct":    ir.BindJSON,
	"google.protobuf.Value":     ir.BindJSON,
	"google.protobuf.ListValue": ir.BindJSON,
	"google.protobuf.Any":       ir.BindJSON,
	"google.protobuf.Empty":     ir.BindJSON,
}

var scalarBinds = map[protoreflect.Kind]ir.Bind{
	protoreflect.BoolKind:     ir.BindBool,
	protoreflect.StringKind:   ir.BindString,
	protoreflect.BytesKind:    ir.BindString,
	protoreflect.Int32Kind:    ir.BindInt,
	protoreflect.Sint32Kind:   ir.BindInt,
	protoreflect.Sfixed32Kind: ir.BindInt,
	protoreflect.Int64Kind:    ir.BindInt,
	protoreflect.Sint64Kind:   ir.BindInt,
	protoreflect.Sfixed64Kind: ir.BindInt,
	protoreflect.Uint32Kind:   ir.BindUint,
	protoreflect.Fixed32Kind:  ir.BindUint,
	protoreflect.Uint64Kind:   ir.BindUint,
	protoreflect.Fixed64Kind:  ir.BindUint,
	protoreflect.FloatKind:    ir.BindFloat,
	protoreflect.DoubleKind:   ir.BindFloat,
}

// buildFlags derives a command's flags from its request message's fields.
func buildFlags(md protoreflect.MessageDescriptor, opts Options) []*ir.Flag {
	var flags []*ir.Flag
	var walk func(md protoreflect.MessageDescriptor, protoPrefix, cliPrefix string, budget int, ancestors []protoreflect.FullName)
	walk = func(md protoreflect.MessageDescriptor, protoPrefix, cliPrefix string, budget int, ancestors []protoreflect.FullName) {
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			path := protoPrefix + string(fd.Name())
			name := cliPrefix + strcase.KebabCase(string(fd.Name()))
			bind, ok, enumValues := fieldBind(fd)

			var oneof string
			if oo := fd.ContainingOneof(); oo != nil && !oo.IsSynthetic() {
				oneof = protoPrefix + string(oo.Name())
			}

			if !ok {
				continue
			}

			if slices.Contains(reservedFlagNames, name) {
				opts.Warn(fmt.Sprintf(
					"field %s.%s: --%s is reserved by a global flag; its flag is --arg-%s",
					md.Name(),
					fd.Name(),
					name,
					name,
				))
				name = "arg-" + name
			}

			flags = append(flags,
				&ir.Flag{
					ProtoPath:  path,
					Name:       name,
					Bind:       bind,
					Repeated:   fd.IsList(),
					Map:        fd.IsMap(),
					EnumValues: enumValues,
					Oneof:      oneof,
				})

			if isExpandable(fd) && budget > 0 &&
				!slices.Contains(ancestors, fd.Message().FullName()) {
				walk(fd.Message(), path+".", name+".", budget-1,
					append(ancestors, fd.Message().FullName()))
			}
		}
	}

	walk(md, "", "", opts.RequestExpandDepth, []protoreflect.FullName{md.FullName()})

	return flags
}

// ok is false for fields with no flag.
func fieldBind(fd protoreflect.FieldDescriptor) (bind ir.Bind, ok bool, enum []string) {
	// A map field's own kind is its synthetic entry message.
	elem := fd
	if fd.IsMap() {
		elem = fd.MapValue()
	}
	switch elem.Kind() {
	case protoreflect.MessageKind:
		if bind, ok = messageBinds[elem.Message().FullName()]; !ok {
			bind, ok = ir.BindJSON, true
		}
	case protoreflect.EnumKind:
		bind, ok = ir.BindString, true
		values := elem.Enum().Values()
		for i := range values.Len() {
			enum = append(enum, string(values.Get(i).Name()))
		}
	default:
		bind, ok = scalarBinds[elem.Kind()]
	}
	return bind, ok, enum
}
