package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	lmb "github.com/suzuki-shunsuke/lambuild/pkg/lambda"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
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
