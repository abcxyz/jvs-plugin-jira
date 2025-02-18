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
	"errors"
	"fmt"
	"net/url"
	"strconv"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jvspb "github.com/abcxyz/jvs/apis/v0"
)

const (
	// jiraCategory is the justification category this plugin will be validating.
	jiraCategory = "jira"

	// JiraIssueID is the key for the Jira Issue ID in the annotation map of the justification.
	jiraIssueID = "jira_issue_id"

	// jiraIssueURL is the key for the Jira Issue URL in the annotation map of the justification.
	jiraIssueURL = "jira_issue_url"
)

// issueMatcher is the mockable interface for the convenience of testing.
type issueMatcher interface {
	MatchIssue(context.Context, string) (*MatchResult, error)
}

// JiraPlugin is the implementation of jvspb.Validator interface.
type JiraPlugin struct {
	validator    issueMatcher
	uiData       *jvspb.UIData
	issueBaseURL string
}

// NewJiraPlugin creates a new JiraPlugin.
func NewJiraPlugin(ctx context.Context, cfg *PluginConfig) (*JiraPlugin, error) {
	apiToken, err := secretVersion(ctx, cfg.APITokenSecretID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch API token: %w", err)
	}

	v, err := NewValidator(cfg.JIRAEndpoint, cfg.Jql, cfg.JIRAAccount, apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate validator: %w", err)
	}

	d := &jvspb.UIData{
		DisplayName: cfg.DisplayName,
		Hint:        cfg.Hint,
	}

	return &JiraPlugin{
		validator:    v,
		uiData:       d,
		issueBaseURL: cfg.IssueBaseURL,
	}, nil
}

// Validate returns the validation result.
func (j *JiraPlugin) Validate(ctx context.Context, req *jvspb.ValidateJustificationRequest) (*jvspb.ValidateJustificationResponse, error) {
	if got, want := req.GetJustification().GetCategory(), jiraCategory; got != want {
		return invalidErrResponse(fmt.Sprintf("failed to perform validation, expected category %q to be %q", got, want)), nil
	}

	if req.GetJustification().GetValue() == "" {
		return invalidErrResponse("empty justification value"), nil
	}

	result, err := j.validateWithJiraEndpoint(ctx, req.GetJustification().GetValue())
	if err != nil {
		if errors.Is(err, errInvalidJustification) {
			return invalidErrResponse(err.Error()),
				nil
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	issueID := strconv.Itoa(result.MatchedIssues[0])
	// The format for the Jira issue URL follows the pattern "https://your-domain.atlassian.net/browse/<issueKey>".
	issueURL, err := url.JoinPath(j.issueBaseURL, "browse", req.GetJustification().GetValue())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &jvspb.ValidateJustificationResponse{
		Valid:   true,
		Warning: result.Errors,
		Annotation: map[string]string{
			jiraIssueID:  issueID,
			jiraIssueURL: issueURL,
		},
	}, nil
}

// Validates the justification with the jira endpoint.
// TODO(#46): move this function to j.validator.MatchIssue.
func (j *JiraPlugin) validateWithJiraEndpoint(ctx context.Context, justificationValue string) (*Match, error) {
	result, err := j.validator.MatchIssue(ctx, justificationValue)
	if err != nil {
		return nil, fmt.Errorf("failed to match jira issue with justification %q: %w", justificationValue, err)
	}

	if len(result.Matches) == 0 || len(result.Matches[0].MatchedIssues) == 0 {
		return nil, fmt.Errorf("no matched jira issue for justification %q: %w", justificationValue, errInvalidJustification)
	}

	// There is only one JQL and one issueKey, only one matching result is expected.
	if len(result.Matches[0].MatchedIssues) > 1 {
		return nil, fmt.Errorf("ambiguous justification %q, multiple matching jira issues are found %v: %w", justificationValue, result.Matches[0].MatchedIssues, errInvalidJustification)
	}

	return result.Matches[0], nil
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

func invalidErrResponse(errStr string) *jvspb.ValidateJustificationResponse {
	return &jvspb.ValidateJustificationResponse{
		Valid: false,
		Error: []string{errStr},
	}
}
