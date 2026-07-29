package irbuild

import (
	cliv0 "github.com/braveokafor/proto-to-cli/proto/cli/v0"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func serviceOptions(sd protoreflect.ServiceDescriptor) *cliv0.ServiceOptions {
	return proto.GetExtension(sd.Options(), cliv0.E_Service).(*cliv0.ServiceOptions)
}

func commandOptions(md protoreflect.MethodDescriptor) *cliv0.CommandOptions {
	return proto.GetExtension(md.Options(), cliv0.E_Command).(*cliv0.CommandOptions)
}

func paramOptions(fd protoreflect.FieldDescriptor) *cliv0.ParamOptions {
	return proto.GetExtension(fd.Options(), cliv0.E_Param).(*cliv0.ParamOptions)
}

func viewOptions(md protoreflect.MessageDescriptor) *cliv0.ViewOptions {
	return proto.GetExtension(md.Options(), cliv0.E_View).(*cliv0.ViewOptions)
}
