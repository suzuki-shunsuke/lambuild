package lambda

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"gopkg.in/yaml.v2"
)

func (handler *Handler) readAppConfig(ctx context.Context, cfg *config.Config) error {
	appName := os.Getenv("APPCONFIG_APPLICATION_NAME")
	if appName == "" {
		return errors.New(`APPCONFIG_APPLICATION_NAME is required`)
	}
	appEnv := os.Getenv("APPCONFIG_ENVIRONMENT_NAME")
	if appEnv == "" {
		return errors.New(`APPCONFIG_ENVIRONMENT_NAME is required`)
	}
	appCfgName := os.Getenv("APPCONFIG_CONFIG_NAME")
	if appCfgName == "" {
		return errors.New(`APPCONFIG_CONFIG_NAME is required`)
	}
	endpoint := "http://localhost:2772/applications/" + appName + "/environments/" + appEnv + "/configurations/" + appCfgName
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create a HTTP request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { //nolint:gomnd
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("HTTP status code >= 300 (%d): failed to read response body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("HTTP status code >= 300 (%d): %s", resp.StatusCode, string(b))
	}
	if err := yaml.NewDecoder(resp.Body).Decode(cfg); err != nil {
		return fmt.Errorf("decode response body as YAML: %w", err)
	}
	logrus.Info("read configuration from AppConfig")
	return nil
}
