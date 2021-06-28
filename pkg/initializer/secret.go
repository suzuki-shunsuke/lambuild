package initializer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
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

func readSecretFromSecretsManager(ctx context.Context, svc *secretsmanager.SecretsManager, secretConfig config.SecretsManager) (lambda.Secret, error) {
	ret := lambda.Secret{}
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretConfig.SecretID),
	}
	if secretConfig.VersionID != "" {
		input.VersionId = aws.String(secretConfig.VersionID)
	}
	output, err := svc.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return ret, fmt.Errorf("get secret value from AWS SecretsManager: %w", err)
	}
	if err := json.Unmarshal([]byte(*output.SecretString), &ret); err != nil {
		return ret, fmt.Errorf("parse secret value: %w", err)
	}
	return ret, nil
}
