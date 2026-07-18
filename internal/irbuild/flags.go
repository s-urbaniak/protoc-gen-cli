package irbuild

import (
	"fmt"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// buildFlags derives a command's flags from its request message's fields.
func buildFlags(md protoreflect.MessageDescriptor, warn func(string)) []*ir.Flag {
	var flags []*ir.Flag
	fields := md.Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)

		bind, ok := scalarBinds[fd.Kind()]
		if !ok || fd.IsList() {
			continue
		}

		name := strcase.KebabCase(string(fd.Name()))
		if reservedFlagNames[name] {
			warn(fmt.Sprintf(
				"field %s.%s: the derived flag --%s is already reserved by a global flag; no flag generated, set the field via -f/-i, or rename it",
				md.Name(),
				fd.Name(),
				name,
			))
			continue
		}

		f := &ir.Flag{
			ProtoPath: string(fd.Name()),
			Name:      name,
			Bind:      bind,
		}

		flags = append(flags, f)
	}

	return flags
}

// Flag names the generated code already reserve.
var reservedFlagNames = map[string]bool{"filename": true, "input": true, "help": true}

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
