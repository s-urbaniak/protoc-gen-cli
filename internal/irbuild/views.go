package irbuild

import (
	"slices"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/tidwall/gjson"
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
			// A json_name may contain gjson syntax characters; escape them so
			// the path addresses the emitted key literally.
			jsonPath := jsonPrefix + gjson.Escape(fd.JSONName())

			if top && fd.IsList() && fd.Kind() == protoreflect.MessageKind {
				if _, wellKnown := messageBinds[fd.Message().FullName()]; !wellKnown {
					view.Lists = append(view.Lists, &ir.ViewList{
						FullName: string(fd.Message().FullName()),
						Label:    string(fd.Name()),
						Path:     gjson.Escape(fd.JSONName()),
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
			cols = append(cols, &ir.ViewField{Label: protoPath, Path: jsonPath})
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
