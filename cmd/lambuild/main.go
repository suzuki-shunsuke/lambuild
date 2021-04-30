package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Webhook struct {
	Body    string  `json:"body"`
	Headers Headers `json:"headers"`
}

type Body struct {
	Ref        string     `json:"ref"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	Sender     Sender     `json:"sender"`
	HeadCommit HeadCommit `json:"head_commit"`
}

type Headers struct {
	Event     string `json:"x-github-event"`
	Delivery  string `json:"x-github-delivery"`
	Signature string `json:"x-hub-signature-256"`
}

type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type Sender struct {
	Login string `json:"login"`
}

type Pusher struct {
	Name string `json:"name"`
}

type HeadCommit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type Handler struct {
	Config               Config
	Secret               Secret
	SecretID             string
	SecretVersionID      string
	Region               string
	BuildspecS3Bucket    string
	BuildspecS3KeyPrefix string
	GitHub               *github.Client
	CodeBuild            *codebuild.CodeBuild
	S3Uploader           *s3manager.Uploader
}

type Secret struct {
	GitHubToken   string `json:"github_token"`
	WebhookSecret string `json:"webhook_secret"`
}

func (handler *Handler) Do(ctx context.Context, webhook Webhook) error {
	if err := github.ValidateSignature(webhook.Headers.Signature, []byte(webhook.Body), []byte(handler.Secret.WebhookSecret)); err != nil {
		// TODO return 400
		fmt.Println(err)
		return nil
	}
	body := Body{}
	if err := json.Unmarshal([]byte(webhook.Body), &body); err != nil {
		return err
	}

	for _, repo := range handler.Config.Repositories {
		if repo.Name != body.Repository.FullName {
			continue
		}
		for _, hook := range repo.Hooks {
			if hook.Event != webhook.Headers.Event {
				continue
			}
			// TODO filter refs
			// TODO get config from github
			owner := strings.Split(body.Repository.FullName, "/")[0]
			path := "lambuild.yaml"
			file, _, _, err := handler.GitHub.Repositories.GetContents(ctx, owner, body.Repository.Name, path, nil)
			if err != nil {
				// start build
				return nil
			}
			content, err := file.GetContent()
			if err != nil {
				return fmt.Errorf("get a content: %w", err)
			}
			// TODO generate buildspec
			buildspecKey := handler.BuildspecS3KeyPrefix + webhook.Headers.Delivery + "/buildspec.yml"
			uploadOut, err := handler.S3Uploader.UploadWithContext(ctx, &s3manager.UploadInput{
				Bucket: aws.String(handler.BuildspecS3Bucket),
				Key:    aws.String(buildspecKey),
				Body:   strings.NewReader(content),
			})
			if err != nil {
				return fmt.Errorf("upload a buildspec: %w", err)
			}
			buildOut, err := handler.CodeBuild.StartBuildBatchWithContext(ctx, &codebuild.StartBuildBatchInput{
				BuildspecOverride: aws.String(uploadOut.Location),
				ProjectName:       aws.String(repo.CodeBuild.ProjectName),
				SourceVersion:     aws.String(body.After),
			})
			if err != nil {
				return fmt.Errorf("start a batch build: %w", err)
			}
			log.Printf("start a build (repo: %s, arn: %s)", body.Repository.FullName, *buildOut.BuildBatch.Arn)
			return nil
		}
	}
	return nil
}

type Config struct {
	Repositories []RepositoryConfig
}

type RepositoryConfig struct {
	Name      string
	Hooks     []HookConfig
	CodeBuild CodeBuildConfig `yaml:"codebuild"`
}

type CodeBuildConfig struct {
	ProjectName string `yaml:"project_name"`
}

type HookConfig struct {
	Event  string
	Refs   string
	Config string
}

func (handler *Handler) Init(ctx context.Context) error {
	cfg := Config{}
	configRaw := os.Getenv("CONFIG")
	if err := yaml.Unmarshal([]byte(configRaw), &cfg); err != nil {
		return err
	}
	handler.Config = cfg
	handler.Region = os.Getenv("REGION")
	handler.BuildspecS3Bucket = os.Getenv("BUILDSPEC_S3_BUCKET")
	handler.BuildspecS3KeyPrefix = os.Getenv("BUILDSPEC_S3_KEY_PREFIX")

	sess := session.Must(session.NewSession())
	if err := handler.readSecret(ctx, sess); err != nil {
		return err
	}
	handler.GitHub = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: handler.Secret.GitHubToken},
	)))
	handler.CodeBuild = codebuild.New(sess, aws.NewConfig().WithRegion(handler.Region))
	handler.S3Uploader = s3manager.NewUploader(sess)

	return nil
}

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	ctx := context.Background()
	handler := Handler{}
	if err := handler.Init(ctx); err != nil {
		return err
	}
	lambda.Start(handler.Do)
	return nil
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
