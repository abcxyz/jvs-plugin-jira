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
	"strconv"

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

// JiraPlugin is the implementation of jvspb.Validator interface.
type JiraPlugin struct {
	validator issueMatcher
	uiData    *jvspb.UIData
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

	d := &jvspb.UIData{
		DisplayName: cfg.DisplayName,
		Hint:        cfg.Hint,
	}

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

	// There is only one JQL and one issueKey, so the first match result is
	// checked directly.
	if len(result.Matches[0].MatchedIssues) == 1 {
		issueID := strconv.Itoa(result.Matches[0].MatchedIssues[0])
		baseURL := "https://verily-okta-sandbox.atlassian.net/browse/"
		issueURL := baseURL + req.Justification.Value
		return &jvspb.ValidateJustificationResponse{
			Valid:   true,
			Warning: result.Matches[0].Errors,
			Annotation: map[string]string{
				"jira_issue_id":  issueID,
				"jira_issue_url": issueURL,
			},
		}, nil
	}

	return &jvspb.ValidateJustificationResponse{
		Valid:   false,
		Warning: result.Matches[0].Errors,
	}, nil
}

func (j *JiraPlugin) GetUIData(ctx context.Context, req *jvspb.GetUIDataRequest) (*jvspb.UIData, error) {
	return j.uiData, nil
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
