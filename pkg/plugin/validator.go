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

// Package validator provides functions to validate jira issue against
// validation criteria.
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	// jiraResponseSizeLimitBytes is the maximum bytes be read from JIRA REST
	// API response.
	jiraResponseSizeLimitBytes = 4_000_000 // 4mb
)

// Validator validates jira issue against validation criteria.
type Validator struct {
	// baseURL is the JIRA REST API url. Example:
	//     https://your-domain.atlassian.net/rest/api/3
	baseURL *url.URL

	// httpClient is an HTTP client used for making outbound requests.
	httpClient *http.Client

	// account is the user name used in [JIRA Basic Auth].
	//
	// [JIRA Basic Auth]: https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
	account string

	// apiToken is the API token used in [JIRA Basic Auth].
	//
	// [JIRA Basic Auth]: https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
	apiToken string

	// jql is the [JQL] query specifying validation criteria.
	//
	// [JQL]: https://support.atlassian.com/jira-service-management-cloud/docs/use-advanced-search-with-jira-query-language-jql/
	jql string
}

// jiraIssue is the representation of a [jira issue].
//
// [jira issue]: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-issueidorkey-get
type jiraIssue struct {
	Key string `json:"key"`
	ID  string `json:"id"`
}

// matchData contains data needed in the request body of a [match request].
//
// [match request]: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-jql-match-post
type matchData struct {
	IssueIDs []string `json:"issueIds"`
	Jqls     []string `json:"jqls"`
}

// Match reports a single match result of the [match request].
//
// [match request]: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-jql-match-post
type Match struct {
	MatchedIssues []int    `json:"matchedIssues"`
	Errors        []string `json:"errors"`
}

// MatchResult reports full list of result of the [match request].
//
// [match request]: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-jql-match-post
type MatchResult struct {
	Matches []*Match `json:"matches"`
}

// NewValidator creates a new
func NewValidator(baseURL, jql, account, apiToken string) (*Validator, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseURL %s: %w", baseURL, err)
	}
	return &Validator{
		baseURL:    u,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		jql:        jql,
		account:    account,
		apiToken:   apiToken,
	}, nil
}

// MatchIssue checks the jira issue against the JQL criteria.
func (v *Validator) MatchIssue(ctx context.Context, issueKey string) (*MatchResult, error) {
	issue, err := v.jiraIssue(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get jira issue %q: %w", issueKey, err)
	}

	result, err := v.matchJQL(ctx, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to validate jira issue %q: %w", issueKey, err)
	}
	return result, nil
}

// jiraIssue sends a request to jira endpoint and returns the jira issue.
func (v *Validator) jiraIssue(ctx context.Context, issueIDOrKey string) (*jiraIssue, error) {
	// Construct [Get Issue API].
	//
	// [Get Issue API]: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-issueidorkey-get
	u := &url.URL{
		Scheme: v.baseURL.Scheme,
		Host:   v.baseURL.Host,
		Path:   path.Join(v.baseURL.Path, "issue", issueIDOrKey),
	}

	q := u.Query()
	q.Set("fields", "key,id")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct jira issue request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	var jiraIssue jiraIssue
	if err := v.makeRequest(req, &jiraIssue); err != nil {
		return nil, err
	}

	return &jiraIssue, nil
}

// matchJQL checks the jira issue against the JQL.
func (v *Validator) matchJQL(ctx context.Context, issue *jiraIssue) (*MatchResult, error) {
	// Construct [Match API].
	//
	// [Match API]: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-jql-match-post
	u := &url.URL{
		Scheme: v.baseURL.Scheme,
		Host:   v.baseURL.Host,
		Path:   path.Join(v.baseURL.Path, "jql", "match"),
	}

	// Create the request body.
	data := matchData{
		IssueIDs: []string{issue.ID},
		Jqls:     []string{v.jql},
	}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	var result MatchResult
	if err := v.makeRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// makeRequest sends an HTTP request, decodes the response and stores the data
// in the value pointed by respVal.
func (v *Validator) makeRequest(req *http.Request, respVal any) error {
	req.SetBasicAuth(v.account, v.apiToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		// Return ErrInternal if jira api returns http status code 5xx.
		return fmt.Errorf(
			"failed to make request to %s, got response code %d: %w",
			req.URL.String(), resp.StatusCode, err)
	} else if resp.StatusCode >= http.StatusBadRequest {
		// Return errInvalidJustification if jira api returns http status code 4xx.
		return fmt.Errorf(
			"failed to make request to %s, got response code %d: %w",
			req.URL.String(), resp.StatusCode, errors.Join(errInvalidJustification, err))
	}

	r := io.LimitReader(resp.Body, jiraResponseSizeLimitBytes)
	if err := json.NewDecoder(r).Decode(&respVal); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
