package irbuild

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	cliv0 "github.com/braveokafor/proto-to-cli/proto/cli/v0"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// buildView builds the view of a message. An annotated message gets its
// declared fields. Other messages get derived fields.
func buildView(md protoreflect.MessageDescriptor, opts Options) *ir.View {
	view := &ir.View{FullName: string(md.FullName())}

	var walk func(md protoreflect.MessageDescriptor, protoPrefix, jsonPrefix string, budget int, ancestors []protoreflect.FullName, top bool) []*ir.ViewField
	walk = func(md protoreflect.MessageDescriptor, protoPrefix, jsonPrefix string, budget int, ancestors []protoreflect.FullName, top bool) []*ir.ViewField {
		var cols []*ir.ViewField
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			protoPath := protoPrefix + string(fd.Name())
			// A json_name can contain gjson syntax characters. Escape them so
			// that the path points to the emitted key exactly.
			jsonPath := jsonPrefix + gjson.Escape(fd.JSONName())

			if top && fd.IsList() && fd.Kind() == protoreflect.MessageKind {
				if _, wellKnown := messageBinds[fd.Message().FullName()]; !wellKnown {
					// This field becomes a sub-table below, not a view field.
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

	// The declared (cli.v0.view) fields of a message replace its derived
	// fields in all of its views. Its expand_depth tunes the derived walk.
	fieldsFor := func(md protoreflect.MessageDescriptor, top bool) []*ir.ViewField {
		vo := viewOptions(md)
		if declared := vo.GetFields(); len(declared) > 0 {
			return declaredFields(md, declared, opts.Warn)
		}
		budget := opts.ResponseExpandDepth
		if vo.HasExpandDepth() {
			budget = int(vo.GetExpandDepth())
		}
		return walk(md, "", "", budget, []protoreflect.FullName{md.FullName()}, top)
	}

	view.Fields = fieldsFor(md, true)

	// Top-level repeated message fields become sub-tables of their element.
	fields := md.Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		if !fd.IsList() || fd.Kind() != protoreflect.MessageKind {
			continue
		}
		if _, wellKnown := messageBinds[fd.Message().FullName()]; wellKnown {
			continue
		}
		view.Lists = append(view.Lists, &ir.ViewList{
			FullName: string(fd.Message().FullName()),
			Label:    string(fd.Name()),
			Path:     gjson.Escape(fd.JSONName()),
			Fields:   fieldsFor(fd.Message(), false),
		})
	}

	return view
}

func declaredFields(
	md protoreflect.MessageDescriptor,
	declared []*cliv0.ViewField,
	warn func(string),
) []*ir.ViewField {
	var cols []*ir.ViewField
	for _, f := range declared {
		path, ok := viewPath(md, f.GetPath())
		if !ok {
			warn(fmt.Sprintf(
				"message %s: view path %q does not resolve; the declared field is dropped",
				md.FullName(),
				f.GetPath(),
			))
			continue
		}
		cols = append(cols, &ir.ViewField{Label: cmp.Or(f.GetLabel(), f.GetPath()), Path: path})
	}
	return cols
}

// viewPath makes the protojson path for a dotted proto-name path. The path
// can go through only singular message fields that expand.
func viewPath(md protoreflect.MessageDescriptor, path string) (string, bool) {
	var parts []string
	for _, seg := range strings.Split(path, ".") {
		if md == nil {
			return "", false
		}
		fd := md.Fields().ByName(protoreflect.Name(seg))
		if fd == nil {
			return "", false
		}
		parts = append(parts, gjson.Escape(fd.JSONName()))
		md = nil
		if isExpandable(fd) {
			md = fd.Message()
		}
	}
	return strings.Join(parts, "."), true
}
