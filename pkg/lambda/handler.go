package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/go-github/v35/github"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Handler struct {
	Config               config.Config
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

func (handler *Handler) Do(ctx context.Context, webhook wh.Webhook) error {
	if err := github.ValidateSignature(webhook.Headers.Signature, []byte(webhook.Body), []byte(handler.Secret.WebhookSecret)); err != nil {
		// TODO return 400
		fmt.Println(err)
		return nil
	}
	body := wh.Body{}
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

func (handler *Handler) Init(ctx context.Context) error {
	cfg := config.Config{}
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
