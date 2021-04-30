package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type Secret struct {
	GitHubToken   string `json:"github_token"`
	WebhookSecret string `json:"webhook_secret"`
}

func (handler *Handler) readSecret(ctx context.Context, sess *session.Session) error {
	svc := ssm.New(sess, aws.NewConfig().WithRegion(handler.Region))
	var err error

	handler.Secret.GitHubToken, err = handler.getSecret(ctx, svc, "github_token")
	if err != nil {
		return fmt.Errorf("get github_token: %w", err)
	}

	handler.Secret.WebhookSecret, err = handler.getSecret(ctx, svc, "lambuild_webhook_secret")
	if err != nil {
		return fmt.Errorf("get a secret webhook: %w", err)
	}

	return nil
}

func (handler *Handler) getSecret(ctx context.Context, svc *ssm.SSM, key string) (string, error) {
	out, err := svc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("get secret value from AWS Systems Manager Parameter Store: %w", err)
	}
	return *out.Parameter.Value, nil
}

// func (handler *Handler) readSecret(ctx context.Context, sess *session.Session) error {
// 	svc := secretsmanager.New(sess, aws.NewConfig().WithRegion(handler.Region))
// 	input := &secretsmanager.GetSecretValueInput{
// 		SecretId: aws.String(handler.SecretID),
// 	}
// 	if handler.SecretVersionID != "" {
// 		input.VersionId = aws.String(handler.SecretVersionID)
// 	}
// 	output, err := svc.GetSecretValueWithContext(ctx, input)
// 	if err != nil {
// 		return fmt.Errorf("get secret value from AWS SecretsManager: %w", err)
// 	}
// 	secret := Secret{}
// 	if err := json.Unmarshal([]byte(*output.SecretString), &secret); err != nil {
// 		return fmt.Errorf("parse secret value: %w", err)
// 	}
// 	handler.Secret = secret
// 	return nil
// }
