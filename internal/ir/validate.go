package ir

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Validate returns every violation found in the model, joined into one error.
func (m *Model) Validate() error {
	var errs []error

	if m.ProtoFile == "" {
		errs = append(errs, errors.New("missing proto_file"))
	}

	for _, svc := range m.Services {
		if len(svc.Commands) == 0 {
			errs = append(errs, fmt.Errorf(
				"service %s has no rpcs to generate commands from; add an rpc or exclude the file from generation",
				svc.ProtoName,
			))
		}

		for _, cmd := range svc.Commands {
			errs = append(errs, validateCommand(svc, cmd)...)
		}
	}

	return errors.Join(errs...)
}

func validateCommand(svc *Service, cmd *Command) []error {
	name := fmt.Sprintf("%s.%s", svc.ProtoName, cmd.ProtoName)
	var errs []error

	// validName reports whether s can name a flag.
	validName := func(s string) bool {
		return s != "" && !strings.HasPrefix(s, "-") && !strings.ContainsFunc(s, unicode.IsSpace)
	}

	flagNames := map[string]string{}
	for _, f := range cmd.Flags {
		prior, taken := flagNames[f.Name]

		if !validName(f.Name) {
			errs = append(errs, fmt.Errorf(
				"rpc %s: field %q derives an unusable flag name %q (a flag name must be nonempty, with no spaces and no leading '-'); rename the field",
				name,
				f.ProtoPath,
				f.Name,
			))
		}

		if taken {
			errs = append(errs, fmt.Errorf(
				"rpc %s: fields %q and %q both derive the flag --%s; rename one of the fields",
				name, prior, f.ProtoPath, f.Name))
		}

		flagNames[f.Name] = f.ProtoPath
	}

	return errs
}
