package bigcli

import (
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"golang.org/x/exp/maps"
	"golang.org/x/xerrors"
)

// Disable is a sentinel value for Option.Flag and Option.Env to disable
// features.
const Disable = "-"

// Annotations is an arbitrary key-mapping used to extend the Option type.
// Its methods won't panic if the map is nil.
type Annotations map[string]string

// Mark sets a value on the attonations map, creating one
// if it doesn't exist. Mark does not mutate the original and
// returns a copy. It is suitable for chaining.
func (a Annotations) Mark(key string, value string) Annotations {
	var aa Annotations
	if a != nil {
		aa = maps.Clone(a)
	} else {
		aa = make(Annotations)
	}
	aa[key] = value
	return aa
}

// IsSet returns true if the key is set in the annotations map.
func (a Annotations) IsSet(key string) bool {
	if a == nil {
		return false
	}
	_, ok := a[key]
	return ok
}

// Get retrieves a key from the map, returning false if the key is not found
// or the map is nil.
func (a Annotations) Get(key string) (string, bool) {
	if a == nil {
		return "", false
	}
	v, ok := a[key]
	return v, ok
}

// Option is a configuration option for a CLI application.
type Option struct {
	Name        string
	Description string

	// If unset, Flag defaults to the kebab-case version of Name.
	// Use sentinel value `Disable` to disable flag support.
	Flag string

	FlagShorthand string

	// If unset, Env defaults to the upper-case, snake-case version of Name.
	// Use special value "Disable" to disable environment variable support.
	Env string

	// Default is parsed into Value if set.
	Default string

	// Value includes the types listed in values.go.
	Value pflag.Value

	// Annotations enable extensions to bigcli higher up in the stack. It's useful for
	// help formatting and documentation generation.
	Annotations Annotations

	// UseInstead is a list of options that should be used instead of this one.
	// The field is used to generate a deprecation warning.
	UseInstead []Option

	Hidden bool
}

// FlagName returns the flag name for the option.
func (o *Option) FlagName() (string, bool) {
	if o.Flag == Disable {
		return "", false
	}
	if o.Flag == "" {
		return strcase.ToKebab(o.Name), true
	}
	return o.Flag, true
}

// EnvName returns the environment variable name for the option.
func (o *Option) EnvName() (string, bool) {
	if o.Env == Disable {
		return "", false
	}
	if o.Env != "" {
		return o.Env, true
	}
	return strings.ToUpper(strcase.ToSnake(o.Name)), true
}

// OptionSet is a group of options that can be applied to a command.
type OptionSet []Option

// Add adds the given Options to the OptionSet.
func (s *OptionSet) Add(opts ...Option) {
	*s = append(*s, opts...)
}

// FlagSet returns a pflag.FlagSet for the OptionSet.
func (s *OptionSet) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	for _, opt := range *s {
		flagName, ok := opt.FlagName()
		if !ok {
			continue
		}

		// HACK: allow omitting value for boolean flags.
		var noOptDefValue string
		{
			no, ok := opt.Value.(NoOptDefValuer)
			if ok {
				noOptDefValue = no.NoOptDefValue()
			}
		}

		fs.AddFlag(&pflag.Flag{
			Name:        flagName,
			Shorthand:   opt.FlagShorthand,
			Usage:       opt.Description,
			Value:       opt.Value,
			DefValue:    "",
			Changed:     false,
			Deprecated:  "",
			NoOptDefVal: noOptDefValue,
			Hidden:      opt.Hidden,
		})
	}
	fs.Usage = func() {
		_, _ = os.Stderr.WriteString("Override (*FlagSet).Usage() to print help text.\n")
	}
	return fs
}

// ParseEnv parses the given environment variables into the OptionSet.
func (s *OptionSet) ParseEnv(globalPrefix string, environ []string) error {
	var merr *multierror.Error

	// We parse environment variables first instead of using a nested loop to
	// avoid N*M complexity when there are a lot of options and environment
	// variables.
	envs := make(map[string]string)
	for _, env := range environ {
		env = strings.TrimPrefix(env, globalPrefix)
		if len(env) == 0 {
			continue
		}

		tokens := strings.SplitN(env, "=", 2)
		if len(tokens) != 2 {
			return xerrors.Errorf("invalid env %q", env)
		}
		envs[tokens[0]] = tokens[1]
	}

	for _, opt := range *s {
		envName, ok := opt.EnvName()
		if !ok {
			continue
		}

		envVal, ok := envs[envName]
		if !ok {
			continue
		}

		if err := opt.Value.Set(envVal); err != nil {
			merr = multierror.Append(
				merr, xerrors.Errorf("parse %q: %w", opt.Name, err),
			)
		}
	}

	return merr.ErrorOrNil()
}

// SetDefaults sets the default values for each Option.
// It should be called before all parsing (e.g. ParseFlags, ParseEnv).
func (os *OptionSet) SetDefaults() error {
	var merr *multierror.Error
	for _, opt := range *os {
		if opt.Default == "" {
			continue
		}
		if opt.Value == nil {
			merr = multierror.Append(
				merr,
				xerrors.Errorf(
					"parse %q: no Value field set\nFull opt: %+v",
					opt.Name, opt,
				),
			)
			continue
		}
		if err := opt.Value.Set(opt.Default); err != nil {
			merr = multierror.Append(
				merr, xerrors.Errorf("parse %q: %w", opt.Name, err),
			)
		}
	}
	return merr.ErrorOrNil()
}
