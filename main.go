package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
	cluster := flag.String("cluster", "", "ECS cluster name")
	service := flag.String("service", "", "ECS service name")
	container := flag.String("container", "", "container name")
	region := flag.String("region", "ap-northeast-1", "AWS Region")
	command := flag.String("command", "/bin/sh", "execute command")
	profile := flag.String("profile", "", "aws profile")

	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(*region), config.WithSharedConfigProfile(*profile))
	if err != nil {
		log.Fatalf("設定情報の読み込みに失敗しました。: %w", err)
	}
	client := ecs.NewFromConfig(cfg)

	taskID, err := getTaskID(client, *cluster, *service)
	if err != nil {
		log.Fatalf("task ID の取得に失敗しました。: %w", err)
	}

	runtimeID, err := getRuntimeID(client, taskID, *cluster, *container)
	if err != nil {
		log.Fatalf("runtime ID の取得に失敗しました。: %w", err)
	}

	resp, err := client.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
		Cluster:     aws.String(*cluster),
		Command:     aws.String(*command),
		Container:   aws.String(*container),
		Interactive: true,
		Task:        aws.String(taskID),
	})

	target := fmt.Sprintf("ecs:%s_%s_%s", *cluster, taskID, runtimeID)

	err = startSession(resp.Session, *region, target)
}

func getTaskID(client *ecs.Client, cluster, service string) (string, error) {
	resp, err := client.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	})
	if err != nil {
		return "", err
	}
	taskArns := resp.TaskArns
	if len(taskArns) == 0 {
		return "", errors.New("TaskArnsが取得できませんでした")
	}
	taskID := strings.Split(taskArns[0], "/")[2]
	return taskID, nil
}

func getRuntimeID(client *ecs.Client, taskID, cluster, container string) (string, error) {
	descTasks, err := client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Tasks:   []string{taskID},
		Cluster: aws.String(cluster),
	})
	if err != nil {
		return "", err
	}
	var runtimeID string
	for _, c := range descTasks.Tasks[0].Containers {
		if *c.Name == container {
			runtimeID = strings.Split(*c.RuntimeId, "-")[0]
		}
	}
	return runtimeID, nil
}

func startSession(sess *types.Session, region string, target string) error {
	sessJSON, _ := json.Marshal(sess)
	endpoint := fmt.Sprintf("https://ssm.%s.amazonaws.com", region)
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
	signal.Ignore(syscall.SIGINT)
	cmd.Run()
	return nil
}
