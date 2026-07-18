package gocobra

import (
	"text/template"

	"github.com/braveokafor/proto-to-cli/internal/ir"
)

// funcs is the helper-function map the template renders with.
var funcs = template.FuncMap{
	"goRPCName": func(svc *ir.Service, cmd *ir.Command) string {
		return svc.ProtoName + cmd.ProtoName
	},
	"goClientType": func(svc *ir.Service) string { return svc.ProtoName + "Client" },
}
