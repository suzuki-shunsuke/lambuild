package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/lambuild/pkg/initializer"
	lmb "github.com/suzuki-shunsuke/lambuild/pkg/lambda"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.WithFields(logrus.Fields{
		"version":        version,
		"built_revision": commit,
		"built_date":     date,
	}).Info("start program")
	if err := core(); err != nil {
		logrus.Fatal(err)
	}
}

func setLogLevel() error {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("parse LOG_LEVEL (%s): %w", logLevel, err)
		}
		logrus.SetLevel(lvl)
	}
	return nil
}

func core() error {
	if err := setLogLevel(); err != nil {
		return fmt.Errorf("set a log level: %w", err)
	}
	ctx := context.Background()
	handler := lmb.Handler{}
	if err := initializer.InitializeHandler(ctx, &handler); err != nil {
		return fmt.Errorf("initialize the Lambda Function: %w", err)
	}
	logrus.Debug("start handler")
	lambda.Start(handler.Do)
	return nil
}
