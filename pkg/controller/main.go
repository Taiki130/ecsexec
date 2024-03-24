package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Controller struct {
	client *ecs.Client
}

func New(client *ecs.Client) *Controller {
	return &Controller{
		client: client,
	}
}

func (ctrl *Controller) Run(ctx context.Context, region, cluster, taskID, container, runtimeID, command string) error {
	resp, err := executeCommand(ctx, ctrl.client, cluster, taskID, container, command)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	target := fmt.Sprintf("ecs:%s_%s_%s", cluster, taskID, runtimeID)
	err = startSession(resp.Session, region, target)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	return nil
}

func LoadConfig(cfg aws.Config) *ecs.Client {
	return ecs.NewFromConfig(cfg)
}

func executeCommand(ctx context.Context, client *ecs.Client, cluster, taskID, container, command string) (*ecs.ExecuteCommandOutput, error) {
	return client.ExecuteCommand(ctx, &ecs.ExecuteCommandInput{
		Cluster:     aws.String(cluster),
		Command:     aws.String(command),
		Container:   aws.String(container),
		Interactive: true,
		Task:        aws.String(taskID),
	})
}

func startSession(sess *types.Session, region string, target string) error {
	sessJSON, _ := json.Marshal(sess)
	endpoint := getSSMEndpoint(region)
	payload := ssm.StartSessionInput{
		Target: aws.String(target),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"session-manager-plugin", string(sessJSON), region, "StartSession", "", string(payloadJSON), endpoint,
	)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	return nil
}

func getSSMEndpoint(region string) string {
	return fmt.Sprintf("https://ssm.%s.amazonaws.com", region)
}
