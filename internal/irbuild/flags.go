package irbuild

import (
	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// buildFlags flattens md into one command's flags.
func buildFlags(md protoreflect.MessageDescriptor) []*ir.Flag {
	var flags []*ir.Flag
	fields := md.Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)

		bind, ok := scalarBinds[fd.Kind()]
		if !ok || fd.IsList() {
			continue
		}

		f := &ir.Flag{
			ProtoPath: string(fd.Name()),
			Name:      strcase.KebabCase(string(fd.Name())),
			Bind:      bind,
		}

		flags = append(flags, f)
	}

	return flags
}

// scalarLeaves maps every proto scalar kind to its JSON Bind.
var scalarBinds = map[protoreflect.Kind]ir.Bind{
	protoreflect.BoolKind:     ir.BindBool,
	protoreflect.StringKind:   ir.BindString,
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
