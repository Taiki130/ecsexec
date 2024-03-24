package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
)

func LoadConfig(cfg aws.Config) *ecs.Client {
	return ecs.NewFromConfig(cfg)
}

func SelectCluster(ctx context.Context, client *ecs.Client) (string, error) {
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

func SelectService(ctx context.Context, client *ecs.Client, cluster string) (string, error) {
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

func GetTaskID(ctx context.Context, client *ecs.Client, cluster, service string) (string, error) {
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

func ExecuteCommand(ctx context.Context, client *ecs.Client, cluster, taskID, container, command string) (*ecs.ExecuteCommandOutput, error) {
	return client.ExecuteCommand(ctx, &ecs.ExecuteCommandInput{
		Cluster:     aws.String(cluster),
		Command:     aws.String(command),
		Container:   aws.String(container),
		Interactive: true,
		Task:        aws.String(taskID),
	})
}

func StartSession(sess *types.Session, region string, target string) error {
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

func PromptRegion() (string, error) {
	l := "Enter Region"
	prompt := promptui.Prompt{
		Label: l,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func SelectProfile() (string, error) {
	l := "Select profile"

	fname := config.DefaultSharedConfigFilename()
	profiles, err := getProfilesFromIni(fname)
	if err != nil {
		return "", err
	}

	prompt := promptui.Select{
		Label: l,
		Items: profiles,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func getProfilesFromIni(fname string) (profiles []string, err error) {
	f, err := ini.Load(fname)
	if err != nil {
		return profiles, err
	}

	for _, v := range f.Sections() {
		if len(v.Keys()) != 0 {
			profile := getProfileFromIniSection(v.Name())
			profiles = append(profiles, profile)
		}
	}
	return
}

func getProfileFromIniSection(section string) string {
	return strings.Split(section, " ")[1]
}
