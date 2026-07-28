package irbuild

import (
	cliv0 "github.com/braveokafor/proto-to-cli/proto/cli/v0"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// serviceOptions reads (cli.v0.service).
func serviceOptions(sd protoreflect.ServiceDescriptor) *cliv0.ServiceOptions {
	return proto.GetExtension(sd.Options(), cliv0.E_Service).(*cliv0.ServiceOptions)
}

// commandOptions reads (cli.v0.command).
func commandOptions(md protoreflect.MethodDescriptor) *cliv0.CommandOptions {
	return proto.GetExtension(md.Options(), cliv0.E_Command).(*cliv0.CommandOptions)
}
