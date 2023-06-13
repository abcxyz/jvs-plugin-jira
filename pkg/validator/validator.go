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

package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	jiraResponseSizeLimitInBytes = 4_000_000 // 4mb
)

type Validator struct {
	// Baseurl of the jira API
	// Example https://your-domain.atlassian.net/rest/api/3
	baseURL string

	// httpClient is an HTTP client used for making outbound requests.
	httpClient *http.Client
}

type JiraIssue struct {
	Key string `json:"key"`
	ID  string `json:"id"`
}

type Match struct {
	MatchedIssues []int    `json:"matchedIssues"`
	Errors        []string `json:"errors"`
}

type MatchResult struct {
	Matches []Match `json:"matches"`
}

type matchData struct {
	IssueIds []string `json:"issueIds"`
	Jqls     []string `json:"jqls"`
}

func NewValidator(url string) (*Validator, error) {
	return &Validator{
		baseURL:    url,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (v *Validator) MatchIssue(ctx context.Context, issueKey, jql, jiraAccount, apiToken string) (*MatchResult, error) {
	issue, err := v.jiraIssue(ctx, issueKey, jiraAccount, apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get jira issue %s: %w", issueKey, err)
	}

	result, err := v.matchJQL(ctx, issue, jql, jiraAccount, apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate jira issue %s: %w", issueKey, err)
	}
	return result, nil
}

func (v *Validator) jiraIssue(ctx context.Context, issueIDOrKey, jiraAccount, apiToken string) (*JiraIssue, error) {
	// Construct get issue api.
	// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-issueidorkey-get
	u, err := url.Parse(fmt.Sprintf("%s/issue/%s?fields=key,id", v.baseURL, issueIDOrKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse getIssue url: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	header := map[string]string{"Accept": "application/json"}

	var jiraIssue JiraIssue
	if err := v.makeRequest(req, jiraAccount, apiToken, header, &jiraIssue); err != nil {
		return nil, err
	}

	return &jiraIssue, nil
}

func (v *Validator) matchJQL(ctx context.Context, issue *JiraIssue, jql, jiraAccount, apiToken string) (*MatchResult, error) {
	// Construct match api.
	// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-jql-match-post

	u, err := url.Parse(fmt.Sprintf("%s/jql/match", v.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse getIssue url: %w", err)
	}

	// Create the request body.
	data := matchData{
		IssueIds: []string{issue.ID},
		Jqls:     []string{jql},
	}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	header := map[string]string{"Accept": "application/json", "Content-Type": "application/json"}

	var result MatchResult
	if err := v.makeRequest(req, jiraAccount, apiToken, header, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Make request and store the data in the value pointed by respVal.
func (v *Validator) makeRequest(req *http.Request, jiraAccount, apiToken string, header map[string]string, respVal any) error {
	req.SetBasicAuth(jiraAccount, apiToken)
	for k, val := range header {
		req.Header.Set(k, val)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	r := io.LimitReader(resp.Body, jiraResponseSizeLimitInBytes)
	if err := json.NewDecoder(r).Decode(&respVal); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
