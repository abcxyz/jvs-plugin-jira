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
	"log"
	"net/http"
)

type Validator struct {
	// Baseurl of the jira API
	// Example https://your-domain.atlassian.net/rest/api/3
	baseURL string
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

func (v *Validator) MatchIssue(ctx context.Context, issueKey, jql, jiraAccount, apiToken string) (*MatchResult, error) {
	issue, err := v.getJiraIssue(ctx, issueKey, jiraAccount, apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get jira issue %s: %w", issueKey, err)
	}

	result, err := v.matchJQL(ctx, *issue, jql, jiraAccount, apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate jira issue %s: %w", issueKey, err)
	}
	return result, nil
}

func (v *Validator) getJiraIssue(ctx context.Context, issueIDOrKey, jiraAccount, apiToken string) (*JiraIssue, error) {
	// Construct get issue api.
	// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-issueidorkey-get
	url := fmt.Sprintf("%s/issue/%s?fields=key,id", v.baseURL, issueIDOrKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	req.SetBasicAuth(jiraAccount, apiToken)
	req.Header.Set("Accept", "application/json")

	var jiraIssue JiraIssue
	v.makeRequest(req, &jiraIssue)

	return &jiraIssue, nil
}

func (v *Validator) matchJQL(ctx context.Context, issue JiraIssue, jql, jiraAccount, apiToken string) (*MatchResult, error) {
	// Construct match api.
	// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-jql-match-post

	url := fmt.Sprintf("%s/jql/match", v.baseURL)

	// Create the request body.
	data := make(map[string][]string)
	data["issueIds"] = append(data["issueIds"], issue.ID)
	data["jqls"] = append(data["jqls"], jql)
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	req.SetBasicAuth(jiraAccount, apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	var result MatchResult
	if err := v.makeRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Make request and store the data in the value pointed by respVal.
func (v *Validator) makeRequest(req *http.Request, respVal any) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	jsonBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, respVal); err != nil {
		return fmt.Errorf("failed to parse request: %w", err)
	}

	return nil
}
