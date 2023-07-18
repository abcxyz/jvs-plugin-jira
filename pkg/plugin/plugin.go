// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package plugin provides the implementation of the JVS plugin interface.
package plugin

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/abcxyz/jvs-plugin-jira/pkg/validator"
	jvspb "github.com/abcxyz/jvs/apis/v0"
)

const (
	// jiraCategory is the justification category this plugin will be validating.
	jiraCategory = "jira"
)

// issueMatcher is the mockable interface for the convenience of testing.
type issueMatcher interface {
	MatchIssue(context.Context, string) (*validator.MatchResult, error)
}

// The JiraUIData comprises the data that will be displayed. At present, it exclusively includes the displayName and hint.
type JiraUIData struct {
	displayName string
	hint        string
}

// JiraPlugin is the implementation of jvspb.Validator interface.
type JiraPlugin struct {
	validator issueMatcher
	uiData    JiraUIData
}

// NewJiraPlugin creates a new JiraPlugin.
func NewJiraPlugin(ctx context.Context, cfg *PluginConfig) (*JiraPlugin, error) {
	apiToken, err := secretVersion(ctx, cfg.APITokenSecretID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch API token: %w", err)
	}

	v, err := validator.NewValidator(cfg.JIRAEndpoint, cfg.Jql, cfg.JIRAAccount, apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate validator: %w", err)
	}

	d := GetUIDataFromPluginConfig(cfg)

	return &JiraPlugin{
		validator: v,
		uiData:    d,
	}, nil
}

// Validate returns the validation result.
func (j *JiraPlugin) Validate(ctx context.Context, req *jvspb.ValidateJustificationRequest) (*jvspb.ValidateJustificationResponse, error) {
	if got, want := req.Justification.Category, jiraCategory; got != want {
		return nil,
			fmt.Errorf("failed to perform validation, expected category %q to be %q",
				got, want)
	}

	if req.Justification.Value == "" {
		return nil, fmt.Errorf("empty justification value")
	}

	result, err := j.validator.MatchIssue(ctx, req.Justification.Value)
	if err != nil {
		return &jvspb.ValidateJustificationResponse{
			Valid: false,
			Error: []string{err.Error()},
		}, fmt.Errorf("failed to validate justification: %w", err)
	}

	return &jvspb.ValidateJustificationResponse{
		// There is only one JQL and one issueKey, so the first match result is
		// checked directly.
		Valid:   len(result.Matches[0].MatchedIssues) == 1,
		Warning: result.Matches[0].Errors,
	}, nil
}

func (j *JiraPlugin) GetUIData(ctx context.Context, req *jvspb.GetUIDataRequest) (*jvspb.UIData, error) {
	return &jvspb.UIData{
		DisplayName: j.uiData.displayName,
		Hint:        j.uiData.hint,
	}, nil
}

// secretVersion returns the secret data as a string.
func secretVersion(ctx context.Context, secretVersionName string) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to set up secret manager client: %w", err)
	}
	defer client.Close()

	// Fetch secret version.
	resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretVersionName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to access API token from secret manager: %w", err)
	}

	return string(resp.GetPayload().GetData()), nil
}

func GetUIDataFromPluginConfig(cfg *PluginConfig) JiraUIData {
	d := cfg.DisplayName
	if d == "" {
		// Sets default category name to display.
		d = "Jira Issue Key"
	}

	h := cfg.Hint
	if h == "" {
		// Sets default hint to display.
		h = "Jira Issue key under JVS project"
	}

	return JiraUIData{
		displayName: d,
		hint:        h,
	}
}
