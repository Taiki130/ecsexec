package main

import (
	"context"

	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/sirupsen/logrus"

	"github.com/Taiki130/ecsexec/pkg/cli"
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
			profileVar, err := cli.SelectProfile()
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
			regionVar, err := cli.PromptRegion()
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

	client := cli.New(cfg)

	if *cluster == "" {
		clusterVar, err := cli.SelectCluster(client)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Faild to retrieve cluster name")
		}
		*cluster = clusterVar
	}

	if *service == "" {
		serviceVar, err := cli.SelectService(client, *cluster)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"error":   err,
				"cluster": *cluster,
			}).Fatal("Failed to retrieve service name")
		}
		*service = serviceVar
	}

	taskID, err := cli.GetTaskID(client, *cluster, *service)

	if err != nil {
		logE.WithFields(logrus.Fields{
			"error":   err,
			"cluster": *cluster,
			"service": *service,
		}).Fatal("Failed to retrieve task ID")
	}

	if *container == "" {
		containerVar, err := cli.SelectContainer(client, *cluster, taskID)
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

	runtimeID, err := cli.GetRuntimeID(client, taskID, *cluster, *container)

	if err != nil {
		logE.WithFields(logrus.Fields{
			"error":     err,
			"cluster":   *cluster,
			"service":   *service,
			"taskID":    taskID,
			"container": *container,
		}).Fatal("Failed to retrieve runtime ID")
	}

	resp, err := cli.Execute(client, *cluster, taskID, *container, *command)

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

	err = cli.StartSession(resp.Session, *region, target)

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
