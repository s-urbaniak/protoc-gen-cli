package ir

import (
	"errors"
	"fmt"
	"strings"
)

// Validate returns every violation found in the model, joined into one error.
func (m *Model) Validate() error {
	var errs []error

	services := map[string]string{}
	for _, svc := range m.Services {
		if prior, taken := services[svc.Name]; taken {
			errs = append(errs, fmt.Errorf(
				"services %s and %s both derive the command %q; rename one of the services",
				prior, svc.ProtoName, svc.Name))
		}
		services[svc.Name] = svc.ProtoName

		if len(svc.Commands) == 0 {
			errs = append(errs, fmt.Errorf(
				"service %s has no rpcs to generate commands from; add an rpc or exclude the file from generation",
				svc.ProtoName,
			))
		}

		commands := map[string]string{}
		for _, cmd := range svc.Commands {
			if prior, taken := commands[cmd.Name]; taken {
				errs = append(errs, fmt.Errorf(
					"rpcs %s.%s and %s.%s both derive the subcommand %q; rename one of the rpcs",
					svc.ProtoName, prior, svc.ProtoName, cmd.ProtoName, cmd.Name))
			}
			commands[cmd.Name] = cmd.ProtoName

			errs = append(errs, validateCommand(svc, cmd)...)
		}
	}

	return errors.Join(errs...)
}

func validateCommand(svc *Service, cmd *Command) []error {
	name := fmt.Sprintf("%s.%s", svc.ProtoName, cmd.ProtoName)
	var errs []error

	// Targets drop '.' and '_' when deriving identifiers, so "a.b_c" and
	// "a_b.c" collide.
	fold := func(path string) string {
		return strings.ToLower(strings.Map(func(r rune) rune {
			if r == '.' || r == '_' {
				return -1
			}
			return r
		}, path))
	}

	seen := map[string]string{}
	folded := map[string]*Param{}
	for _, f := range cmd.Params {
		if prior, taken := seen[f.Name]; taken {
			errs = append(errs, fmt.Errorf(
				"rpc %s: fields %q and %q both derive the flag --%s; rename one of the fields",
				name, prior, f.ProtoPath, f.Name))
		}

		if prior, taken := folded[fold(f.ProtoPath)]; taken && prior.Name != f.Name {
			errs = append(errs, fmt.Errorf(
				"rpc %s: fields %q and %q collapse to the same generated identifier; rename one of the fields",
				name,
				prior.ProtoPath,
				f.ProtoPath,
			))
		}

		seen[f.Name] = f.ProtoPath
		folded[fold(f.ProtoPath)] = f
	}

	return errs
}
