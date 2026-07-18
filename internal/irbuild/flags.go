package irbuild

import (
	"fmt"
	"slices"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// buildFlags derives a command's flags from its request message's fields.
func buildFlags(md protoreflect.MessageDescriptor, opts Options) []*ir.Flag {
	var flags []*ir.Flag
	isWellKnown := func(md protoreflect.MessageDescriptor) bool {
		return md.FullName().Parent() == "google.protobuf"
	}
	var walk func(md protoreflect.MessageDescriptor, protoPrefix, cliPrefix string, budget int, ancestors []protoreflect.FullName)
	walk = func(md protoreflect.MessageDescriptor, protoPrefix, cliPrefix string, budget int, ancestors []protoreflect.FullName) {
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			path := protoPrefix + string(fd.Name())
			name := cliPrefix + strcase.KebabCase(string(fd.Name()))

			bind, ok := scalarBinds[fd.Kind()]
			if fd.Kind() == protoreflect.MessageKind {
				bind, ok = messageBinds[fd.Message().FullName()]
			}

			switch {
			case fd.IsList(), fd.IsMap():
			case ok && reservedFlagNames[name]:
				opts.Warn(fmt.Sprintf(
					"field %s.%s: the derived flag --%s is already reserved by a global flag; no flag generated, set the field via -f/-i, or rename it",
					md.Name(),
					fd.Name(),
					name,
				))
			case ok:
				flags = append(flags, &ir.Flag{ProtoPath: path, Name: name, Bind: bind})
			case fd.Kind() != protoreflect.MessageKind:
			case isWellKnown(fd.Message()):
				// protojson gives every google.protobuf message a special
				// JSON form; only a messageBinds row carries one in a flag.
			case slices.Contains(ancestors, fd.Message().FullName()):
				// A type already being expanded: descending again would only
				// repeat its flags until the budget ran out.
			case budget > 0:
				walk(fd.Message(), path+".", name+".", budget-1,
					append(ancestors, fd.Message().FullName()))
			}
		}
	}
	walk(md, "", "", opts.RequestExpandDepth, []protoreflect.FullName{md.FullName()})

	return flags
}

// Flag names the generated code already reserve.
var reservedFlagNames = map[string]bool{"filename": true, "input": true, "help": true}

// Message types protojson renders as a single JSON value;
// - A Timestamp is an RFC 3339 string
// - A timestamp Duration is a "30s"-style seconds, etc.
var messageBinds = map[protoreflect.FullName]ir.Bind{
	"google.protobuf.Timestamp": ir.BindString,
	"google.protobuf.Duration":  ir.BindString,
	"google.protobuf.FieldMask": ir.BindString,
}

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
