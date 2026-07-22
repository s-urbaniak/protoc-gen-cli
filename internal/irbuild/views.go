package irbuild

import (
	"slices"
	"strings"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// buildView derives a message's view from its fields.
func buildView(md protoreflect.MessageDescriptor, opts Options) *ir.View {
	view := &ir.View{FullName: string(md.FullName())}
	var walk func(md protoreflect.MessageDescriptor, protoPrefix, jsonPrefix string, budget int, ancestors []protoreflect.FullName, top bool) []*ir.ViewField
	walk = func(md protoreflect.MessageDescriptor, protoPrefix, jsonPrefix string, budget int, ancestors []protoreflect.FullName, top bool) []*ir.ViewField {
		var cols []*ir.ViewField
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			protoPath := protoPrefix + string(fd.Name())
			jsonPath := jsonPrefix + fd.JSONName()

			if top && fd.IsList() && fd.Kind() == protoreflect.MessageKind {
				if _, wellKnown := messageBinds[fd.Message().FullName()]; !wellKnown {
					view.Lists = append(view.Lists, &ir.ViewList{
						Label: strings.ToUpper(string(fd.Name())),
						Path:  fd.JSONName(),
						Fields: walk(fd.Message(), "", "", opts.ResponseExpandDepth,
							[]protoreflect.FullName{fd.Message().FullName()}, false),
					})
					continue
				}
			}
			if isExpandable(fd) && budget > 0 &&
				!slices.Contains(ancestors, fd.Message().FullName()) {
				cols = append(cols, walk(fd.Message(), protoPath+".", jsonPath+".", budget-1,
					append(ancestors, fd.Message().FullName()), false)...)
				continue
			}
			cols = append(cols, &ir.ViewField{Label: strings.ToUpper(protoPath), Path: jsonPath})
		}
		return cols
	}

	view.Fields = walk(
		md,
		"",
		"",
		opts.ResponseExpandDepth,
		[]protoreflect.FullName{md.FullName()},
		true,
	)

	return view
}
