package initializer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/lambda"
)

type SSM interface {
	GetParameterWithContext(ctx aws.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error)
}

func readSecretFromSSM(ctx context.Context, svc SSM, parameterName config.ParameterName) (lambda.Secret, error) {
	var err error
	secret := lambda.Secret{}

	secret.GitHubToken, err = getSecret(ctx, svc, parameterName.GitHubToken)
	if err != nil {
		return secret, fmt.Errorf("get GitHub Access Token: %w", err)
	}

	secret.WebhookSecret, err = getSecret(ctx, svc, parameterName.WebhookSecret)
	if err != nil {
		return secret, fmt.Errorf("get a secret webhook: %w", err)
	}

	return secret, nil
}

func getSecret(ctx context.Context, svc SSM, key string) (string, error) {
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
