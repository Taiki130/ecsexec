package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
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
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

func main() {
	cluster := flag.String("cluster", "", "the name of your ECS cluster.")
	service := flag.String("service", "", "the name of your ECS service.")
	container := flag.String("container", "", "the name of container name.")
	region := flag.String("region", "", "AWS Region.")
	command := flag.String("command", "/bin/sh", "Specify the command to execute. Default: /bin/sh")
	profile := flag.String("profile", "", "AWS profile to use.")

	flag.Parse()

	if *profile == "" {
		profileVar, ok := os.LookupEnv("AWS_PROFILE")
		if !ok {
			profileVar, err := selectProfile()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Faild to get profile name")
			}
			*profile = profileVar
		} else {
			*profile = profileVar
		}
	}

	if *region == "" {
		regionVar, ok := os.LookupEnv("AWS_REGION")
		if !ok {
			regionVar, err := promptRegion()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Faild to get region name")
			}
			*region = regionVar
		} else {
			*region = regionVar
		}
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(*region), config.WithSharedConfigProfile(*profile))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to load configuration")
	}

	client := ecs.NewFromConfig(cfg)

	if *cluster == "" {
		clusterVar, err := selectCluster(client)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Faild to retrieve cluster name")
		}
		*cluster = clusterVar
	}

	if *service == "" {
		serviceVar, err := selectService(client, *cluster)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":   err,
				"cluster": *cluster,
			}).Fatal("Failed to retrieve service name")
		}
		*service = serviceVar
	}

	taskID, err := getTaskID(client, *cluster, *service)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"cluster": *cluster,
			"service": *service,
		}).Fatal("Failed to retrieve task ID")
	}

	if *container == "" {
		containerVar, err := selectContainer(client, *cluster, taskID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":   err,
				"cluster": *cluster,
				"service": *service,
				"taskID":  taskID,
			}).Fatal("Faild to retrieve container name")
		}
		*container = containerVar
	}

	runtimeID, err := getRuntimeID(client, taskID, *cluster, *container)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err,
			"cluster":   *cluster,
			"service":   *service,
			"taskID":    taskID,
			"container": *container,
		}).Fatal("Failed to retrieve runtime ID")
	}

	resp, err := client.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
		Cluster:     aws.String(*cluster),
		Command:     aws.String(*command),
		Container:   aws.String(*container),
		Interactive: true,
		Task:        aws.String(taskID),
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err,
			"cluster":   *cluster,
			"service":   *service,
			"taskID":    taskID,
			"container": *container,
		}).Fatal("Failed to execute command")
	}

	target := fmt.Sprintf("ecs:%s_%s_%s", *cluster, taskID, runtimeID)

	err = startSession(resp.Session, *region, target)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err,
			"cluster":   *cluster,
			"service":   *service,
			"taskID":    taskID,
			"container": *container,
		}).Fatal("Session Failed")
	}
}

func selectProfile() (string, error) {
	l := "Select Profile"

	fname := config.DefaultSharedConfigFilename()
	f, err := ini.Load(fname)
	if err != nil {
		return "", err
	}

	var profiles []string
	for _, v := range f.Sections() {
		if len(v.Keys()) != 0 {
			profile := strings.Split(v.Name(), " ")[1]
			profiles = append(profiles, profile)
		}
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

func promptRegion() (string, error) {
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

func selectCluster(client *ecs.Client) (string, error) {
	l := "Select cluster"
	resp, err := client.ListClusters(context.TODO(), &ecs.ListClustersInput{})
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

func selectService(client *ecs.Client, cluster string) (string, error) {
	l := "Select service"
	resp, err := client.ListServices(context.TODO(), &ecs.ListServicesInput{
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

func selectContainer(client *ecs.Client, cluster, taskID string) (string, error) {
	l := "Select container"
	resp, err := client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
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
		return "", errors.New("no ECS task found")
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
