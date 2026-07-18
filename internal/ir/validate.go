package ir

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/stoewer/go-strcase"
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

	// fold collapses a proto path the way targets derive identifiers from
	// it: delimiters vanish, so a.b_c and a_b.c collide.
	fold := func(path string) string {
		var id string
		for seg := range strings.SplitSeq(path, ".") {
			id += strcase.UpperCamelCase(seg)
		}
		return id
	}

	// The command's flag and identifier namespaces. A taken name maps to the
	// proto path that claimed it.
	flagNames := map[string]string{}
	identifiers := map[string]string{}

	for _, f := range cmd.Flags {
		prior, taken := flagNames[f.Name]
		id := fold(f.ProtoPath)
		priorID, idTaken := identifiers[id]

		switch {
		case !validName(f.Name):
			errs = append(errs, fmt.Errorf(
				"rpc %s: field %q derives an unusable flag name %q (a flag name must be nonempty, with no spaces and no leading '-'); rename the field",
				name,
				f.ProtoPath,
				f.Name,
			))
		case taken:
			errs = append(errs, fmt.Errorf(
				"rpc %s: fields %q and %q both derive the flag --%s; rename one of the fields",
				name, prior, f.ProtoPath, f.Name))
		case idTaken:
			errs = append(errs, fmt.Errorf(
				"rpc %s: fields %q and %q derive the same identifier in generated code; rename one of the fields",
				name,
				priorID,
				f.ProtoPath,
			))
		default:
			flagNames[f.Name] = f.ProtoPath
			identifiers[id] = f.ProtoPath
		}
	}

	return errs
}
