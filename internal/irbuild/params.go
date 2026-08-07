package irbuild

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/braveokafor/protoc-gen-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// protojson shows these message types as one JSON value.
var messageBinds = map[protoreflect.FullName]ir.Bind{
	"google.protobuf.Timestamp": ir.BindTimestamp,
	"google.protobuf.Duration":  ir.BindDuration,
	"google.protobuf.FieldMask": ir.BindFieldMask,

	"google.protobuf.BoolValue":   ir.BindBool,
	"google.protobuf.StringValue": ir.BindString,
	"google.protobuf.BytesValue":  ir.BindBytes,
	"google.protobuf.Int32Value":  ir.BindInt,
	"google.protobuf.Int64Value":  ir.BindInt,
	"google.protobuf.UInt32Value": ir.BindUint,
	"google.protobuf.UInt64Value": ir.BindUint,
	"google.protobuf.FloatValue":  ir.BindFloat,
	"google.protobuf.DoubleValue": ir.BindFloat,

	"google.protobuf.Struct":    ir.BindJSON,
	"google.protobuf.Value":     ir.BindJSON,
	"google.protobuf.ListValue": ir.BindList,
	"google.protobuf.Any":       ir.BindAny,
	"google.protobuf.Empty":     ir.BindJSON,
}

var scalarBinds = map[protoreflect.Kind]ir.Bind{
	protoreflect.BoolKind:     ir.BindBool,
	protoreflect.StringKind:   ir.BindString,
	protoreflect.BytesKind:    ir.BindBytes,
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

// buildParams derives a command's params from its request message's fields.
func buildParams(md protoreflect.MessageDescriptor, opts Options) []*ir.Param {
	var params []*ir.Param
	var walk func(md protoreflect.MessageDescriptor, protoPrefix, cliPrefix string, depth int, ancestors []protoreflect.FullName)
	walk = func(md protoreflect.MessageDescriptor, protoPrefix, cliPrefix string, depth int, ancestors []protoreflect.FullName) {
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			po := paramOptions(fd)

			if po.GetSkip() {
				continue
			}

			path := protoPrefix + string(fd.Name())
			name := cliPrefix + cmp.Or(po.GetName(), strcase.KebabCase(string(fd.Name())))
			bind, ok, enumValues := fieldBind(fd)

			var oneof string
			if oo := fd.ContainingOneof(); oo != nil && !oo.IsSynthetic() {
				oneof = protoPrefix + string(oo.Name())
			}

			if !ok {
				opts.Warn(fmt.Sprintf(
					"field %s.%s: group-encoded fields get no flag; set it with -f or -d, or make the field a message",
					md.FullName(),
					fd.Name(),
				))
				continue
			}

			if po.GetName() == "" && slices.Contains(ir.ReservedParamNames, name) {
				opts.Warn(fmt.Sprintf(
					"field %s.%s: --%s is reserved by a built-in flag; its flag is --arg-%s",
					md.FullName(),
					fd.Name(),
					name,
					name,
				))
				name = "arg-" + name
			}

			if !isExpandable(fd) && po.HasExpandDepth() {
				opts.Warn(fmt.Sprintf(
					"field %s.%s: expand_depth is dropped; the field does not expand into sub-flags",
					md.FullName(),
					fd.Name(),
				))
			}

			fieldDepth := depth
			if po.HasExpandDepth() {
				fieldDepth = int(po.GetExpandDepth())
			}
			expands := isExpandable(fd) && fieldDepth > 0 &&
				!slices.Contains(ancestors, fd.Message().FullName())

			if po.GetHoist() && !expands {
				opts.Warn(fmt.Sprintf(
					"field %s.%s: hoist is dropped; the field does not expand into sub-flags",
					md.FullName(),
					fd.Name(),
				))
			}

			comment := fd.ParentFile().SourceLocations().ByDescriptor(fd).LeadingComments
			short, _ := helpFrom(cmp.Or(po.GetHelp(), cleanComment(comment)))
			params = append(params,
				&ir.Param{
					ProtoPath:  path,
					Name:       name,
					Shorthand:  po.GetShorthand(),
					Deprecated: fd.Options().(*descriptorpb.FieldOptions).GetDeprecated(),
					ShortHelp:  short,
					Bind:       bind,
					Repeated:   fd.IsList(),
					Map:        fd.IsMap(),
					EnumValues: enumValues,
					Oneof:      oneof,
				})

			if expands {
				childPrefix := name + "."
				if po.GetHoist() {
					childPrefix = cliPrefix
				}
				walk(fd.Message(), path+".", childPrefix, fieldDepth-1,
					append(ancestors, fd.Message().FullName()))
			}
		}
	}

	walk(md, "", "", opts.RequestExpandDepth, []protoreflect.FullName{md.FullName()})

	return params
}

// fieldBind returns ok false for a field with no param.
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
