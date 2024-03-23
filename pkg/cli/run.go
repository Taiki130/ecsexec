package cli

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/urfave/cli/v2"
)

var err error

func (runner *Runner) execute(ctx *cli.Context) error {
	region := ctx.String("region")
	if region == "" {
		region, err = promptRegion()
		if err != nil {
			return fmt.Errorf("faild to get region name: %w", err)
		}
	}

	profile := ctx.String("profile")
	if profile == "" {
		profile, err = selectProfile()
		if err != nil {
			return fmt.Errorf("faild to get profile name: %w", err)
		}
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region), config.WithSharedConfigProfile(profile))
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := loadConfig(cfg)

	cluster := ctx.String("cluster")
	if cluster == "" {
		cluster, err = selectCluster(client)
		if err != nil {
			return fmt.Errorf("faild to retrieve cluster name: %w", err)
		}
	}

	service := ctx.String("service")
	if service == "" {
		service, err = selectService(client, cluster)
		if err != nil {
			return fmt.Errorf("failed to retrieve service name: %w", err)
		}
	}

	taskID, err := getTaskID(client, cluster, service)
	if err != nil {
		return fmt.Errorf("failed to retrieve task ID: %w", err)
	}

	container := ctx.String("container")
	if container == "" {
		container, err = selectContainer(client, cluster, taskID)
		if err != nil {
			return fmt.Errorf("failed to retrieve container name: %w", err)
		}
	}

	runtimeID, err := getRuntimeID(client, taskID, cluster, container)
	if err != nil {
		return fmt.Errorf("failed to retrieve runtime ID: %w", err)
	}

	command := ctx.String("command")
	if command == "" {
		command = "/bin/bash"
	}

	resp, err := executeCommand(client, cluster, taskID, container, command)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	target := fmt.Sprintf("ecs:%s_%s_%s", cluster, taskID, runtimeID)

	err = startSession(resp.Session, region, target)
	if err != nil {
		return fmt.Errorf("failed session: %w", err)
	}
	return nil
}
