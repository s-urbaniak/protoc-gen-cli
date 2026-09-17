package irbuild

import (
	"encoding/json"
	"testing"

	"github.com/braveokafor/protoc-gen-cli/internal/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	annotations "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestBuildParams_SkipsOutputOnlyFieldsAtEveryDepth(t *testing.T) {
	t.Parallel()

	fieldOptions := func(behaviors ...annotations.FieldBehavior) *descriptorpb.FieldOptions {
		t.Helper()

		opts := &descriptorpb.FieldOptions{}
		proto.SetExtension(opts, annotations.E_FieldBehavior, behaviors)
		return opts
	}

	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test.v1"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Request"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("resource"),
						Number:   proto.Int32(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".test.v1.Resource"),
					},
					{
						Name:    proto.String("server_note"),
						Number:  proto.Int32(2),
						Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Options: fieldOptions(annotations.FieldBehavior_OUTPUT_ONLY),
					},
				},
			},
			{
				Name: proto.String("Resource"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:    proto.String("id"),
						Number:  proto.Int32(1),
						Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Options: fieldOptions(annotations.FieldBehavior_OUTPUT_ONLY),
					},
					{
						Name:    proto.String("reusable"),
						Number:  proto.Int32(2),
						Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:    descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(),
						Options: fieldOptions(annotations.FieldBehavior_IMMUTABLE),
					},
				},
			},
		},
	}, nil)
	require.NoError(t, err)

	params := buildParams(file.Messages().ByName("Request"), Options{
		RequestExpandDepth: 1,
		Warn:               func(string) {},
	})

	assert.Equal(t, []string{"resource", "resource.reusable"}, paramNames(params))
	assert.JSONEq(
		t,
		`{"resource":{"reusable":true}}`,
		buildExample(file.Messages().ByName("Request"), Options{Warn: func(string) {}}),
	)
}

func TestIsOutputOnly_AbsentBehaviorIsFalse(t *testing.T) {
	t.Parallel()

	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test.v1"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Request"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:   proto.String("name"),
				Number: proto.Int32(1),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			}},
		}},
	}, nil)
	require.NoError(t, err)

	assert.False(t, isOutputOnly(file.Messages().ByName("Request").Fields().ByName("name")))
}

func TestBuildExample_OneofSkipsOutputOnlyCandidates(t *testing.T) {
	t.Parallel()

	options := &descriptorpb.FieldOptions{}
	proto.SetExtension(
		options,
		annotations.E_FieldBehavior,
		[]annotations.FieldBehavior{annotations.FieldBehavior_OUTPUT_ONLY},
	)
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test.v1"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name:      proto.String("Request"),
			OneofDecl: []*descriptorpb.OneofDescriptorProto{{Name: proto.String("choice")}},
			Field: []*descriptorpb.FieldDescriptorProto{
				{
					Name:       proto.String("server_choice"),
					Number:     proto.Int32(1),
					Label:      descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:       descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					OneofIndex: proto.Int32(0),
					Options:    options,
				},
				{
					Name:       proto.String("input_choice"),
					Number:     proto.Int32(2),
					Label:      descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:       descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					OneofIndex: proto.Int32(0),
				},
			},
		}},
	}, nil)
	require.NoError(t, err)

	var example map[string]any
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(buildExample(file.Messages().ByName("Request"), Options{Warn: func(string) {}})),
			&example,
		),
	)
	assert.NotContains(t, example, "serverChoice")
	assert.Contains(t, example, "inputChoice")
}

func paramNames(params []*ir.Param) []string {
	names := make([]string, 0, len(params))
	for _, param := range params {
		names = append(names, param.Name)
	}
	return names
}
