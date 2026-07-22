package irbuild

import (
	"slices"
	"strings"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// buildView derives a message's table view from its fields.
func buildView(md protoreflect.MessageDescriptor, budget int) *ir.View {
	var viewFields []*ir.ViewField
	var walk func(md protoreflect.MessageDescriptor, protoPrefix, jsonPrefix string, budget int, ancestors []protoreflect.FullName)
	walk = func(md protoreflect.MessageDescriptor, protoPrefix, jsonPrefix string, budget int, ancestors []protoreflect.FullName) {
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			protoPath := protoPrefix + string(fd.Name())
			jsonPath := jsonPrefix + fd.JSONName()

			if isExpandable(fd) && budget > 0 &&
				!slices.Contains(ancestors, fd.Message().FullName()) {
				walk(fd.Message(), protoPath+".", jsonPath+".", budget-1,
					append(ancestors, fd.Message().FullName()))
				continue
			}

			viewFields = append(viewFields, &ir.ViewField{
				Label: strings.ToUpper(protoPath),
				Path:  jsonPath,
			})
		}
	}

	walk(md, "", "", budget, []protoreflect.FullName{md.FullName()})

	return &ir.View{
		FullName: string(md.FullName()),
		Fields:   viewFields,
	}
}
