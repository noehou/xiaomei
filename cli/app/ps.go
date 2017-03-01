package app

import (
	"fmt"

	"github.com/bughou-go/xiaomei/cli/cluster"
	"github.com/bughou-go/xiaomei/config"
	"github.com/spf13/cobra"
)

func PsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   `ps [<env>]`,
		Short: `list tasks of app service.`,
		RunE: func(c *cobra.Command, args []string) error {
			env := `dev`
			if len(args) > 0 {
				env = args[0]
			}
			return Ps(env)
		},
	}
}

func Ps(env string) error {
	return cluster.Run(env, fmt.Sprintf(`docker service ps %s_app`, config.DeployName()))
}

func Restart(env string) error {
	return nil
}

func Shell(env string) error {
	return nil
}

func Exec(env string) error {
	return nil
}
