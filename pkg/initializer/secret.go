package initializer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/suzuki-shunsuke/lambuild/pkg/lambda"
)

func readSecretFromSSM(ctx context.Context, handler *lambda.Handler, sess *session.Session) error {
	svc := ssm.New(sess, aws.NewConfig().WithRegion(handler.Config.Region))
	var err error

	handler.Secret.GitHubToken, err = getSecret(ctx, svc, handler.Config.SSMParameter.ParameterName.GitHubToken)
	if err != nil {
		return fmt.Errorf("get GitHub Access Token: %w", err)
	}

	handler.Secret.WebhookSecret, err = getSecret(ctx, svc, handler.Config.SSMParameter.ParameterName.WebhookSecret)
	if err != nil {
		return fmt.Errorf("get a secret webhook: %w", err)
	}

	return nil
}

func getSecret(ctx context.Context, svc *ssm.SSM, key string) (string, error) {
	out, err := svc.GetParameterWithContext(ctx, &ssm.GetParameterInput{
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
