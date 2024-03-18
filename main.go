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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/manifoldco/promptui"
)

func main() {
	cluster := flag.String("cluster", "", "the name of your ECS cluster.")
	service := flag.String("service", "", "the name of your ECS service.")
	container := flag.String("container", "", "the name of container name.")
	region := flag.String("region", "", "AWS Region.")
	command := flag.String("command", "/bin/sh", "Specify the command to execute. Default: /bin/sh")
	profile := flag.String("profile", "", "AWS profile to use.")

	flag.Parse()

	if *region == "" {
		regionVar, ok := os.LookupEnv("AWS_REGION")
		if !ok {
			fmt.Print("Enter region: ")
			fmt.Scanf("%s", &region)
		} else {
			*region = regionVar
		}
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(*region), config.WithSharedConfigProfile(*profile))
	if err != nil {
		log.Fatalf("Failed to load configuration: %w", err)
	}
	client := ecs.NewFromConfig(cfg)

	if *cluster == "" {
		selectCluster(client)
	}

	taskID, err := getTaskID(client, *cluster, *service)
	if err != nil {
		log.Fatalf("Failed to retrieve task ID: %w", err)
	}

	runtimeID, err := getRuntimeID(client, taskID, *cluster, *container)
	if err != nil {
		log.Fatalf("Failed to retrieve runtime ID: %w", err)
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

func selectCluster(client *ecs.Client) (string, error) {
	l := "Select cluster"
	resp, err := client.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	if err != nil {
		return "", err
	}
	clusterArns := resp.ClusterArns
	if len(clusterArns) == 0 {
		return "", errors.New("no ECS cluster found.")
	}
	var i []string
	for _, arn := range clusterArns {
		i = append(i, strings.Split(arn, "/")[1])
	}
	prompt := promptui.Select{
		Label: l,
		Items: i,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
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
		return "", errors.New("No ECS task found.")
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
	cmd.Run()
	return nil
}
