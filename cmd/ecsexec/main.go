package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"

	"github.com/Taiki130/ecsexec/pkg/ecs"
	"github.com/Taiki130/ecsexec/pkg/log"
)

func main() {
	logE := log.New()

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
				logE.WithFields(logrus.Fields{
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
				logE.WithFields(logrus.Fields{
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
		logE.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to load configuration")
	}

	client := ecs.New(cfg)

	if *cluster == "" {
		clusterVar, err := ecs.SelectCluster(client)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Faild to retrieve cluster name")
		}
		*cluster = clusterVar
	}

	if *service == "" {
		serviceVar, err := ecs.SelectService(client, *cluster)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"error":   err,
				"cluster": *cluster,
			}).Fatal("Failed to retrieve service name")
		}
		*service = serviceVar
	}

	taskID, err := ecs.GetTaskID(client, *cluster, *service)

	if err != nil {
		logE.WithFields(logrus.Fields{
			"error":   err,
			"cluster": *cluster,
			"service": *service,
		}).Fatal("Failed to retrieve task ID")
	}

	if *container == "" {
		containerVar, err := ecs.SelectContainer(client, *cluster, taskID)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"error":   err,
				"cluster": *cluster,
				"service": *service,
				"taskID":  taskID,
			}).Fatal("Faild to retrieve container name")
		}
		*container = containerVar
	}

	runtimeID, err := ecs.GetRuntimeID(client, taskID, *cluster, *container)

	if err != nil {
		logE.WithFields(logrus.Fields{
			"error":     err,
			"cluster":   *cluster,
			"service":   *service,
			"taskID":    taskID,
			"container": *container,
		}).Fatal("Failed to retrieve runtime ID")
	}

	resp, err := ecs.Execute(client, *cluster, taskID, *container, *command)

	if err != nil {
		logE.WithFields(logrus.Fields{
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
		logE.WithFields(logrus.Fields{
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
