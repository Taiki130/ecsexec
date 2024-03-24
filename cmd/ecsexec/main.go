package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Taiki130/ecsexec/pkg/cli"
	"github.com/Taiki130/ecsexec/pkg/log"
	"github.com/sirupsen/logrus"
)

func main() {
	logE := log.New()
	if err := core(logE); err != nil {
		logE.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("ecsexec failed")
	}
}

func core(logE *logrus.Entry) error {
	runner := cli.Runner{
		LogE: logE,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runner.Run(ctx, os.Args...)
}
