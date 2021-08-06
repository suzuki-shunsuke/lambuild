package initializer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/lambda"
)

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
