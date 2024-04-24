package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func SelectCluster(ctx context.Context, client *ecs.Client) (string, error) {
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

	return Select("cluster", clusterNames)
}

func SelectService(ctx context.Context, client *ecs.Client, cluster string) (string, error) {
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
		var serviceName string
		if strings.Count(arn, "/") == 2 {
			serviceName = strings.Split(arn, "/")[2]
		} else {
			serviceName = strings.Split(arn, "/")[1]
		}
		serviceNames = append(serviceNames, serviceName)
	}

	return Select("service", serviceNames)
}

func SelectTaskIDs(ctx context.Context, client *ecs.Client, cluster, service string) (string, error) {
	resp, err := client.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	})
	if err != nil {
		return "", err
	}
	taskARNs := resp.TaskArns
	if len(taskARNs) == 0 {
		return "", errors.New("no ECS task found")
	}

	var taskIDs []string
	for _, arn := range taskARNs {
		taskIDs = append(taskIDs, getTaskIDFromTaskArn(arn))
	}

	return Select("taskID", taskIDs)
}

func getTaskIDFromTaskArn(taskARN string) string {
	return strings.Split(taskARN, "/")[2]
}

func GetRuntimeID(ctx context.Context, client *ecs.Client, taskID, cluster, container string) (string, error) {
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

func SelectContainer(ctx context.Context, client *ecs.Client, cluster, taskID string) (string, error) {
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

	return Select("container", containerNames)

}

func GetContainers(ctx context.Context, client *ecs.Client, cluster, taskID string) ([]string, error) {
	resp, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskID},
	})
	if err != nil {
		return []string{}, fmt.Errorf("failed to describe tasks: %w", err)
	}

	containers := resp.Tasks[0].Containers
	var containerNames []string
	for _, c := range containers {
		containerName := *c.Name
		containerNames = append(containerNames, containerName)
	}

	return containerNames, nil
}
