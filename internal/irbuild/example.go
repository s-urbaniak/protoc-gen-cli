package irbuild

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// buildExample makes the example request of md, as protojson.
func buildExample(md protoreflect.MessageDescriptor, opts Options) string {
	fail := func() string {
		opts.Warn(fmt.Sprintf(
			"message %s: no example request; --example prints {}, so build it with -f or -i",
			md.FullName(),
		))
		return "{}"
	}

	// A zero seed would reseed gofakeit from crypto/rand on each run.
	faker := gofakeit.New(1)

	var walk func(md protoreflect.MessageDescriptor, ancestors []protoreflect.FullName) map[string]any
	walk = func(md protoreflect.MessageDescriptor, ancestors []protoreflect.FullName) map[string]any {
		value := func(fd protoreflect.FieldDescriptor) (any, bool) {
			if isNested(fd) {
				name := fd.Message().FullName()
				// A recursive field still shows. The reader needs to know it.
				if slices.Contains(ancestors, name) {
					return map[string]any{}, true
				}
				return walk(fd.Message(), append(ancestors, name)), true
			}
			if fd.Kind() == protoreflect.EnumKind {
				return enumExample(fd.Enum())
			}
			bind, ok, _ := fieldBind(fd)
			if !ok {
				return nil, false
			}
			return bindExample(bind, faker)
		}

		// A oneof shows one arm. A synthetic oneof is not a choice.
		arms := map[protoreflect.Name]protoreflect.FieldDescriptor{}
		oneofs := md.Oneofs()
		for i := range oneofs.Len() {
			oo := oneofs.Get(i)
			if oo.IsSynthetic() {
				continue
			}
			if n := oo.Fields().Len(); n > 0 {
				arms[oo.Name()] = oo.Fields().Get(faker.IntN(n))
			}
		}

		doc := map[string]any{}
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			if oo := fd.ContainingOneof(); oo != nil && !oo.IsSynthetic() && arms[oo.Name()] != fd {
				continue
			}

			// A map field's own kind is its synthetic entry message.
			if fd.IsMap() {
				if v, ok := value(fd.MapValue()); ok {
					doc[fd.JSONName()] = map[string]any{exampleKey(fd.MapKey(), faker): v}
				}
				continue
			}
			v, ok := value(fd)
			if !ok {
				continue
			}
			if fd.IsList() {
				v = []any{v}
			}
			doc[fd.JSONName()] = v
		}
		return doc
	}

	doc, err := json.Marshal(walk(md, []protoreflect.FullName{md.FullName()}))
	if err != nil {
		return fail()
	}

	// The round trip validates the example. protojson gives the canonical form.
	msg := dynamicpb.NewMessage(md)
	if err := protojson.Unmarshal(doc, msg); err != nil {
		return fail()
	}
	raw, err := protojson.Marshal(msg)
	if err != nil {
		return fail()
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw)
	}
	return buf.String()
}

// protojson writes every map key as a JSON string.
func exampleKey(fd protoreflect.FieldDescriptor, faker *gofakeit.Faker) string {
	bind, _, _ := fieldBind(fd)
	switch bind {
	case ir.BindBool:
		return "true"
	case ir.BindInt, ir.BindUint:
		return fmt.Sprint(faker.IntRange(1, 1000))
	}
	return faker.Word()
}

// bindExample makes one value of the bind's JSON type. It reports false for a
// field that the example omits.
func bindExample(bind ir.Bind, faker *gofakeit.Faker) (any, bool) {
	switch bind {
	case ir.BindFieldMask, ir.BindAny:
		// Only the server knows these values.
		return nil, false
	case ir.BindBool:
		return true, true
	case ir.BindInt, ir.BindUint, ir.BindFloat:
		// protojson accepts one whole number for every number type.
		return faker.IntRange(1, 1000), true
	case ir.BindBytes:
		return base64.StdEncoding.EncodeToString([]byte(faker.Sentence(3))), true
	case ir.BindTimestamp:
		at := faker.DateRange(
			time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC),
		)
		return at.UTC().Format(time.RFC3339), true
	case ir.BindDuration:
		return fmt.Sprintf("%ds", faker.IntRange(1, 3600)), true
	case ir.BindList:
		return []any{faker.Sentence(3)}, true
	case ir.BindJSON:
		// google.protobuf.Empty accepts no field.
		return map[string]any{}, true
	}
	return faker.Sentence(3), true
}

// enumExample prefers a non-zero value. protojson omits a zero enum.
func enumExample(ed protoreflect.EnumDescriptor) (any, bool) {
	values := ed.Values()
	for i := range values.Len() {
		if v := values.Get(i); v.Number() != 0 {
			return string(v.Name()), true
		}
	}
	return nil, false
}
