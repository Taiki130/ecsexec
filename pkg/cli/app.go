package cli

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	LogE *logrus.Entry
}

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	app := cli.App{
		Name:   "ecsexec",
		Usage:  "Access a shell session within a container running in an ECS task.",
		Action: runner.execute,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "region, r",
				Usage:   "AWS region name.",
				EnvVars: []string{"AWS_REGION"},
			},
			&cli.StringFlag{
				Name:    "profile, p",
				Usage:   "AWS profile name.",
				EnvVars: []string{"AWS_PROFILE"},
			},
			&cli.StringFlag{
				Name:    "cluster, cl",
				Usage:   "ECS cluster name.",
				EnvVars: []string{"ECSEXEC_CLUSTER"},
			},
			&cli.StringFlag{
				Name:    "service, s",
				Usage:   "ECS service name.",
				EnvVars: []string{"ECSEXEC_SERVICE"},
			},
			&cli.StringFlag{
				Name:    "container, co",
				Usage:   "container name.",
				EnvVars: []string{"ECSEXEC_CONTAINER"},
			},
			&cli.StringFlag{
				Name:    "command, cmd",
				Usage:   "login shell. default: /bin/sh",
				EnvVars: []string{"ECSEXEC_COMMAND"},
			},
		},
	}

	return app.RunContext(ctx, args)
}
