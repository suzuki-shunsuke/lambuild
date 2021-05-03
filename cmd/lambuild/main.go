package main

import (
	"context"

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

func core() error {
	ctx := context.Background()
	handler := lmb.Handler{}
	if err := handler.Init(ctx); err != nil {
		return err
	}
	lambda.Start(handler.Do)
	return nil
}
