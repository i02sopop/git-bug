package usercmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/commands/cmdjson"
	"github.com/MichaelMure/git-bug/commands/completion"
	"github.com/MichaelMure/git-bug/commands/execenv"
	"github.com/MichaelMure/git-bug/util/colors"
)

type userOptions struct {
	format string
}

func NewUserCommand(env *execenv.Env) (*cobra.Command, error) {
	options := userOptions{}

	cmd := &cobra.Command{
		Use:     "user",
		Short:   "List identities",
		PreRunE: execenv.LoadBackend(env),
		RunE: execenv.CloseBackend(env, func(cmd *cobra.Command, args []string) error {
			return runUser(env, options)
		}),
	}

	subCmd, err := newUserShowCommand(env)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(newUserNewCommand(env))
	cmd.AddCommand(subCmd)
	cmd.AddCommand(newUserAdoptCommand(env))

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.format, "format", "f", "default",
		"Select the output formatting style. Valid values are [default,json]")
	err = cmd.RegisterFlagCompletionFunc("format", completion.From([]string{"default", "json"}))
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func runUser(env *execenv.Env, opts userOptions) error {
	ids := env.Backend.Identities().AllIds()
	var users []*cache.IdentityExcerpt
	for _, id := range ids {
		user, err := env.Backend.Identities().ResolveExcerpt(id)
		if err != nil {
			return err
		}
		users = append(users, user)
	}

	switch opts.format {
	case "json":
		return userJsonFormatter(env, users)
	case "default":
		return userDefaultFormatter(env, users)
	default:
		return fmt.Errorf("unknown format %s", opts.format)
	}
}

func userDefaultFormatter(env *execenv.Env, users []*cache.IdentityExcerpt) error {
	for _, user := range users {
		env.Out.Printf("%s %s\n",
			colors.Cyan(user.Id().Human()),
			user.DisplayName(),
		)
	}

	return nil
}

func userJsonFormatter(env *execenv.Env, users []*cache.IdentityExcerpt) error {
	jsonUsers := make([]cmdjson.Identity, len(users))
	for i, user := range users {
		jsonUsers[i] = cmdjson.NewIdentityFromExcerpt(user)
	}

	return env.Out.PrintJSON(jsonUsers)
}
