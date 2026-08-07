package irbuild

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/braveokafor/protoc-gen-cli/internal/ir"
	cliv1 "github.com/braveokafor/protoc-gen-cli/proto/cli/v1"
	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// buildView builds the view of a message.
// An annotated message gets its declared fields. Other messages get derived fields.
func buildView(md protoreflect.MessageDescriptor, opts Options) *ir.View {
	view := &ir.View{FullName: string(md.FullName())}

	const skipLists, keepLists = true, false

	var walk func(md protoreflect.MessageDescriptor, protoPrefix string, jsonPrefix []*spec.Segment, depth int, ancestors []protoreflect.FullName, skip bool) []*ir.ViewField
	walk = func(md protoreflect.MessageDescriptor, protoPrefix string, jsonPrefix []*spec.Segment, depth int, ancestors []protoreflect.FullName, skip bool) []*ir.ViewField {
		var cols []*ir.ViewField
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			protoPath := protoPrefix + string(fd.Name())
			jsonPath := append(slices.Clone(jsonPrefix), spec.Child(spec.Name(fd.JSONName())))

			if skip && fd.IsList() && isNested(fd) {
				continue
			}
			if isExpandable(fd) && depth > 0 &&
				!slices.Contains(ancestors, fd.Message().FullName()) {
				cols = append(cols, walk(fd.Message(), protoPath+".", jsonPath, depth-1,
					append(ancestors, fd.Message().FullName()), keepLists)...)
				continue
			}
			cols = append(cols, &ir.ViewField{
				Label: protoPath,
				Path:  jsonpath.New(spec.Query(true, jsonPath...)).String(),
			})
		}
		return cols
	}

	// The declared (cli.v1.view) fields of a message replace its derived
	// fields in all of its views. Its expand_depth tunes the derived walk.
	fieldsFor := func(md protoreflect.MessageDescriptor, skip bool) []*ir.ViewField {
		vo := viewOptions(md)
		if declared := vo.GetFields(); len(declared) > 0 {
			return declaredFields(md, declared, opts.Warn)
		}
		depth := opts.ResponseExpandDepth
		if vo.HasExpandDepth() {
			depth = int(vo.GetExpandDepth())
		}
		return walk(md, "", nil, depth, []protoreflect.FullName{md.FullName()}, skip)
	}

	view.Fields = fieldsFor(md, skipLists)

	fields := md.Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		if !fd.IsList() || !isNested(fd) {
			continue
		}
		// The list path selects the elements, so the fields below read one.
		elements := spec.Query(
			true,
			spec.Child(spec.Name(fd.JSONName())),
			spec.Child(spec.Wildcard()),
		)
		view.Lists = append(view.Lists, &ir.ViewList{
			FullName: string(fd.Message().FullName()),
			Label:    string(fd.Name()),
			Path:     jsonpath.New(elements).String(),
			Fields:   fieldsFor(fd.Message(), keepLists),
		})
	}

	return view
}

func declaredFields(
	md protoreflect.MessageDescriptor,
	declared []*cliv1.ViewField,
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
	var segments []*spec.Segment
	for _, seg := range strings.Split(path, ".") {
		if md == nil {
			return "", false
		}
		fd := md.Fields().ByName(protoreflect.Name(seg))
		if fd == nil {
			return "", false
		}
		segments = append(segments, spec.Child(spec.Name(fd.JSONName())))
		md = nil
		if isExpandable(fd) {
			md = fd.Message()
		}
	}
	return jsonpath.New(spec.Query(true, segments...)).String(), true
}
