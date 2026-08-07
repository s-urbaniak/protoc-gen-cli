package ir

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// The built-in params of every generated command claim these.
var (
	ReservedParamNames = []string{
		"filename",
		"data",
		"help",
		"output",
		"columns",
		"example",
		"timeout",
		"dry-run",
	}
	ReservedShorthands = []string{"f", "d", "h", "o"}
)

type claim struct {
	party string
	alias bool
}

// Validate returns every violation in the model.
func (m *Model) Validate() error {
	var errs []error

	services := map[string]claim{}
	checkedViews := map[string]bool{}
	for _, svc := range m.Services {
		errs = append(errs, claimService(services, svc)...)

		if len(svc.Commands) == 0 {
			errs = append(errs, fmt.Errorf(
				"service %s has no rpcs to generate commands from; add an rpc or exclude the file from generation",
				svc.ProtoName,
			))
		}

		commands := map[string]claim{}
		for _, cmd := range svc.Commands {
			errs = append(errs, claimCommand(commands, svc, cmd)...)
			errs = append(errs, validateCommand(svc, cmd)...)
			errs = append(errs, validateViews(checkedViews, cmd)...)
		}

		errs = append(errs, validateServiceAliases(commands, svc)...)
	}

	return errors.Join(errs...)
}

func claimService(claims map[string]claim, svc *Service) []error {
	var errs []error
	for i, n := range append([]string{svc.Name}, svc.Aliases...) {
		cur := claim{party: svc.ProtoName, alias: i > 0}

		if !validName(n) {
			kind := "command name"
			if cur.alias {
				kind = "alias"
			}
			errs = append(errs, fmt.Errorf(
				"service %s: %s %q must be nonempty, without spaces or a leading '-'",
				svc.ProtoName, kind, n))
		}

		prior, taken := claims[n]
		claims[n] = cur
		switch {
		case !taken:
		case prior.party == cur.party && !prior.alias:
			errs = append(errs, fmt.Errorf(
				"service %s's alias %q duplicates its own command name; drop the alias",
				cur.party, n))
		case prior.party == cur.party:
			errs = append(errs, fmt.Errorf(
				"service %s declares the alias %q twice; drop one",
				cur.party, n))
		case !prior.alias && !cur.alias:
			errs = append(errs, fmt.Errorf(
				"services %s and %s both derive the command %q; rename one of the services",
				prior.party, cur.party, n))
		case prior.alias && cur.alias:
			errs = append(errs, fmt.Errorf(
				"services %s and %s both declare the alias %q; realias one of the services",
				prior.party, cur.party, n))
		default:
			aliased, named := prior.party, cur.party
			if cur.alias {
				aliased, named = cur.party, prior.party
			}
			errs = append(errs, fmt.Errorf(
				"service %s's alias %q collides with the command of service %s; rename or realias one",
				aliased,
				n,
				named,
			))
		}
	}
	return errs
}

func claimCommand(claims map[string]claim, svc *Service, cmd *Command) []error {
	var errs []error
	for i, n := range append([]string{cmd.Name}, cmd.Aliases...) {
		cur := claim{party: cmd.ProtoName, alias: i > 0}

		if !validName(n) {
			kind := "subcommand name"
			if cur.alias {
				kind = "alias"
			}
			errs = append(errs, fmt.Errorf(
				"rpc %s.%s: %s %q must be nonempty, without spaces or a leading '-'",
				svc.ProtoName, cmd.ProtoName, kind, n))
		}

		prior, taken := claims[n]
		claims[n] = cur
		switch {
		case !taken:
		case prior.party == cur.party && !prior.alias:
			errs = append(errs, fmt.Errorf(
				"rpc %s.%s's alias %q duplicates its own subcommand name; drop the alias",
				svc.ProtoName, cur.party, n))
		case prior.party == cur.party:
			errs = append(errs, fmt.Errorf(
				"rpc %s.%s declares the alias %q twice; drop one",
				svc.ProtoName, cur.party, n))
		case !prior.alias && !cur.alias:
			errs = append(errs, fmt.Errorf(
				"rpcs %s.%s and %s.%s both derive the subcommand %q; rename one of the rpcs",
				svc.ProtoName, prior.party, svc.ProtoName, cur.party, n))
		case prior.alias && cur.alias:
			errs = append(errs, fmt.Errorf(
				"rpcs %s.%s and %s.%s both declare the alias %q; realias one of the rpcs",
				svc.ProtoName, prior.party, svc.ProtoName, cur.party, n))
		default:
			aliased, named := prior.party, cur.party
			if cur.alias {
				aliased, named = cur.party, prior.party
			}
			errs = append(errs, fmt.Errorf(
				"rpc %s.%s's alias %q collides with the subcommand of %s.%s; rename or realias one",
				svc.ProtoName, aliased, n, svc.ProtoName, named))
		}
	}
	return errs
}

func validateServiceAliases(commands map[string]claim, svc *Service) []error {
	var errs []error
	seen := map[string]bool{}
	for _, a := range svc.Aliases {
		if seen[a] {
			continue
		}
		seen[a] = true

		c, taken := commands[a]
		switch {
		case !taken:
		case c.alias:
			errs = append(errs, fmt.Errorf(
				"service %s's alias %q collides with an alias of rpc %s.%s; realias one",
				svc.ProtoName, a, svc.ProtoName, c.party))
		default:
			errs = append(errs, fmt.Errorf(
				"service %s's alias %q collides with the subcommand of rpc %s.%s; typing %q would run the command, not the subcommand; realias the service or rename the rpc",
				svc.ProtoName,
				a,
				svc.ProtoName,
				c.party,
				a,
			))
		}
	}
	return errs
}

func validateCommand(svc *Service, cmd *Command) []error {
	name := fmt.Sprintf("%s.%s", svc.ProtoName, cmd.ProtoName)
	var errs []error

	// Targets drop '.' and '_' when they derive identifiers. Thus "a.b_c"
	// and "a_b.c" collide.
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
	shorthands := map[string]string{}
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

		if !validName(f.Name) {
			errs = append(errs, fmt.Errorf(
				"rpc %s: flag name %q must be nonempty, without spaces or a leading '-'",
				name, f.Name))
		}

		if slices.Contains(ReservedParamNames, f.Name) {
			errs = append(errs, fmt.Errorf(
				"rpc %s: flag --%s is reserved by a built-in flag; pick another (cli.v1.param).name",
				name,
				f.Name,
			))
		}

		if f.Shorthand == "" {
			continue
		}
		switch {
		case !validShorthand(f.Shorthand):
			errs = append(errs, fmt.Errorf(
				"rpc %s: flag --%s: shorthand %q must be one ASCII letter",
				name, f.Name, f.Shorthand))
		case slices.Contains(ReservedShorthands, f.Shorthand):
			errs = append(errs, fmt.Errorf(
				"rpc %s: flag --%s: shorthand -%s is reserved by a built-in flag; pick another letter",
				name,
				f.Name,
				f.Shorthand,
			))
		}
		if prior, taken := shorthands[f.Shorthand]; taken {
			errs = append(errs, fmt.Errorf(
				"rpc %s: flags --%s and --%s both claim the shorthand -%s; change one",
				name, prior, f.Name, f.Shorthand))
		}
		shorthands[f.Shorthand] = f.Name
	}

	return errs
}

func validateViews(checked map[string]bool, cmd *Command) []error {
	var errs []error
	check := func(fullName string, fields []*ViewField) {
		if checked[fullName] {
			return
		}
		checked[fullName] = true

		labels := map[string]bool{}
		for _, f := range fields {
			if labels[f.Label] {
				errs = append(errs, fmt.Errorf(
					"message %s: the view declares the label %q twice; relabel one",
					fullName, f.Label))
			}
			labels[f.Label] = true
		}
	}

	if cmd.View == nil {
		return nil
	}
	check(cmd.View.FullName, cmd.View.Fields)
	for _, l := range cmd.View.Lists {
		check(l.FullName, l.Fields)
	}
	return errs
}

func validName(s string) bool {
	return s != "" && !strings.HasPrefix(s, "-") && !strings.ContainsFunc(s, unicode.IsSpace)
}

func validShorthand(s string) bool {
	return len(s) == 1 &&
		('a' <= s[0] && s[0] <= 'z' || 'A' <= s[0] && s[0] <= 'Z')
}
