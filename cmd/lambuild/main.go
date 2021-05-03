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
	setLogLevel()
	if err := core(); err != nil {
		logrus.Fatal(err)
	}
}

func setLogLevel() {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"log_level": logLevel,
			}).WithError(err).Error("the log level is invalid")
		} else {
			logrus.SetLevel(lvl)
		}
	}
}

func core() error {
	ctx := context.Background()
	handler := lmb.Handler{}
	if err := handler.Init(ctx); err != nil {
		return fmt.Errorf("initialize the Lambda Function: %w", err)
	}
	lambda.Start(handler.Do)
	return nil
}
