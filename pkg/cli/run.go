package cli

import (
	"fmt"

	"github.com/Taiki130/ecsexec/pkg/constants"
	"github.com/Taiki130/ecsexec/pkg/controller"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/urfave/cli/v2"
)

var err error

func (runner *Runner) execute(ctx *cli.Context) error {
	region := ctx.String("region")
	if region == "" {
		region, err = controller.Select("region", constants.AWS_VALID_REGIONS)
		if err != nil {
			return fmt.Errorf("failed to get region name: %w", err)
		}
	}

	profile := ctx.String("profile")
	if profile == "" {
		profile, err = controller.SelectProfile()
		if err != nil {
			return fmt.Errorf("faild to get profile name: %w", err)
		}
	}

	cfg, err := config.LoadDefaultConfig(ctx.Context, config.WithRegion(region), config.WithSharedConfigProfile(profile))
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := controller.LoadConfig(cfg)

	cluster := ctx.String("cluster")
	if cluster == "" {
		cluster, err = controller.SelectCluster(ctx.Context, client)
		if err != nil {
			return fmt.Errorf("faild to retrieve cluster name: %w", err)
		}
	}

	service := ctx.String("service")
	if service == "" {
		service, err = controller.SelectService(ctx.Context, client, cluster)
		if err != nil {
			return fmt.Errorf("failed to retrieve service name: %w", err)
		}
	}

	taskID, err := controller.SelectTaskIDs(ctx.Context, client, cluster, service)
	if err != nil {
		return fmt.Errorf("failed to retrieve taskID: %w", err)
	}

	container := ctx.String("container")
	if container == "" {
		container, err = controller.SelectContainer(ctx.Context, client, cluster, taskID)
		if err != nil {
			return fmt.Errorf("failed to retrieve container name: %w", err)
		}
	}

	runtimeID, err := controller.GetRuntimeID(ctx.Context, client, taskID, cluster, container)
	if err != nil {
		return fmt.Errorf("failed to retrieve runtime ID: %w", err)
	}

	command := ctx.String("command")
	if command == "" {
		command = "/bin/sh"
	}

	ctrl := controller.New(client)
	return ctrl.Run(ctx.Context, region, cluster, taskID, container, runtimeID, command)
}
