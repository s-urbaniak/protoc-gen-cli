package irbuild

import (
	cliv1 "github.com/braveokafor/protoc-gen-cli/proto/cli/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func serviceOptions(sd protoreflect.ServiceDescriptor) *cliv1.ServiceOptions {
	return proto.GetExtension(sd.Options(), cliv1.E_Service).(*cliv1.ServiceOptions)
}

func commandOptions(md protoreflect.MethodDescriptor) *cliv1.CommandOptions {
	return proto.GetExtension(md.Options(), cliv1.E_Command).(*cliv1.CommandOptions)
}

func paramOptions(fd protoreflect.FieldDescriptor) *cliv1.ParamOptions {
	return proto.GetExtension(fd.Options(), cliv1.E_Param).(*cliv1.ParamOptions)
}

func viewOptions(md protoreflect.MessageDescriptor) *cliv1.ViewOptions {
	return proto.GetExtension(md.Options(), cliv1.E_View).(*cliv1.ViewOptions)
}
