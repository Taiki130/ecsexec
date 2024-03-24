package cli

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/manifoldco/promptui"
)

func loadConfig(cfg aws.Config) *ecs.Client {
	return ecs.NewFromConfig(cfg)
}

func selectCluster(ctx context.Context, client *ecs.Client) (string, error) {
	l := "Select cluster"
	resp, err := client.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return "", err
	}
	clusterArns := resp.ClusterArns
	if len(clusterArns) == 0 {
		return "", errors.New("no ECS cluster found")
	}

	var clusterNames []string
	for _, arn := range clusterArns {
		clusterName := strings.Split(arn, "/")[1]
		clusterNames = append(clusterNames, clusterName)
	}

	prompt := promptui.Select{
		Label: l,
		Items: clusterNames,
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func selectService(ctx context.Context, client *ecs.Client, cluster string) (string, error) {
	l := "Select service"
	resp, err := client.ListServices(ctx, &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	})
	if err != nil {
		return "", err
	}
	serviceArns := resp.ServiceArns
	if len(serviceArns) == 0 {
		return "", errors.New("no ECS task found")
	}
	var serviceNames []string
	for _, arn := range serviceArns {
		serviceName := strings.Split(arn, "/")[2]
		serviceNames = append(serviceNames, serviceName)
	}
	prompt := promptui.Select{
		Label: l,
		Items: serviceNames,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func getTaskID(ctx context.Context, client *ecs.Client, cluster, service string) (string, error) {
	resp, err := client.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	})
	if err != nil {
		return "", err
	}
	taskArns := resp.TaskArns
	if len(taskArns) == 0 {
		return "", errors.New("no ECS task found")
	}
	taskID := strings.Split(taskArns[0], "/")[2]
	return taskID, nil
}

func getRuntimeID(ctx context.Context, client *ecs.Client, taskID, cluster, container string) (string, error) {
	descTasks, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
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

func selectContainer(ctx context.Context, client *ecs.Client, cluster, taskID string) (string, error) {
	l := "Select container"
	resp, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskID},
	})
	if err != nil {
		return "", err
	}
	containers := resp.Tasks[0].Containers
	var containerNames []string
	for _, c := range containers {
		containerName := *c.Name
		containerNames = append(containerNames, containerName)
	}

	prompt := promptui.Select{
		Label: l,
		Items: containerNames,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, err
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
