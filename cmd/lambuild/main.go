package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	lmb "github.com/suzuki-shunsuke/lambuild/pkg/lambda"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
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
	if err := handler.Init(ctx); err != nil {
		return fmt.Errorf("initialize the Lambda Function: %w", err)
	}
	lambda.Start(handler.Do)
	return nil
}
