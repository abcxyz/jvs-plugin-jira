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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jvspb "github.com/abcxyz/jvs/apis/v0"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
)

type mockValidator struct {
	result *MatchResult
	err    error
}

func (m *mockValidator) MatchIssue(ctx context.Context, issueKey string) (*MatchResult, error) {
	return m.result, m.err
}

func TestPlugin_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		validator *mockValidator
		req       *jvspb.ValidateJustificationRequest
		want      *jvspb.ValidateJustificationResponse
		wantErr   string
	}{
		{
			name: "happy_path",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
					Value:    "ABCD",
				},
			},
			validator: &mockValidator{
				result: &MatchResult{
					Matches: []*Match{
						{
							MatchedIssues: []int{1234},
							Errors:        []string{},
						},
					},
				},
			},
			want: &jvspb.ValidateJustificationResponse{
				Valid:   true,
				Warning: []string{},
				Annotation: map[string]string{
					"jira_issue_id":  "1234",
					"jira_issue_url": "https://example.atlassian.net/browse/ABCD",
				},
			},
		},
		{
			name: "wrong_category",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "github",
					Value:    "ABCD",
				},
			},
			validator: &mockValidator{
				result: &MatchResult{
					Matches: []*Match{
						{
							MatchedIssues: []int{1234},
							Errors:        []string{},
						},
					},
				},
			},
			want: invalidErrResponse("failed to perform validation, expected category \"github\" to be \"jira\""),
		},
		{
			name: "empty_matches",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
					Value:    "ABCD",
				},
			},
			validator: &mockValidator{
				result: &MatchResult{
					Matches: []*Match{},
				},
			},
			want: invalidErrResponse("no matched jira issue for justification \"ABCD\": invalid justification"),
		},
		{
			name: "empty_matchesIssue",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
					Value:    "ABCD",
				},
			},
			validator: &mockValidator{
				result: &MatchResult{
					Matches: []*Match{
						{
							MatchedIssues: []int{},
							Errors:        []string{},
						},
					},
				},
			},
			want: invalidErrResponse("no matched jira issue for justification \"ABCD\": invalid justification"),
		},
		{
			name: "empty_value",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
				},
			},
			validator: &mockValidator{
				result: &MatchResult{
					Matches: []*Match{
						{
							MatchedIssues: []int{1234},
							Errors:        []string{},
						},
					},
				},
			},
			want: invalidErrResponse("empty justification value"),
		},
		{
			name: "not_match_jql",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
					Value:    "ABCD",
				},
			},
			validator: &mockValidator{
				err: fmt.Errorf("non match: %w", errInvalidJustification),
			},
			want: invalidErrResponse("failed to match jira issue with justification \"ABCD\": non match: invalid justification"),
		},
		{
			name: "match_error",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
					Value:    "ABCD",
				},
			},
			validator: &mockValidator{
				err: fmt.Errorf("unexpected error"),
			},
			want:    nil,
			wantErr: status.Errorf(codes.Internal, "failed to match jira issue with justification \"ABCD\": unexpected error").Error(),
		},
		{
			name: "multiple_matches",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
					Value:    "ABCD",
				},
			},
			validator: &mockValidator{
				result: &MatchResult{
					Matches: []*Match{
						{
							MatchedIssues: []int{1234, 5678, 6784},
							Errors:        []string{},
						},
					},
				},
			},
			want: invalidErrResponse("ambiguous justification \"ABCD\", multiple matching jira issues are found [1234 5678 6784]: invalid justification"),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &JiraPlugin{
				validator:    tc.validator,
				issueBaseURL: "https://example.atlassian.net",
			}

			ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))
			got, err := p.Validate(ctx, tc.req)
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreUnexported(jvspb.ValidateJustificationResponse{})); diff != "" {
				t.Errorf("Failed validation (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestPlugin_GetUIData(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		req     *jvspb.GetUIDataRequest
		uiData  *jvspb.UIData
		want    *jvspb.UIData
		wantErr string
	}{
		{
			name: "success",
			req:  &jvspb.GetUIDataRequest{},
			uiData: &jvspb.UIData{
				DisplayName: "Jira Issue key",
				Hint:        "Jira Issue key under JVS project",
			},
			want: &jvspb.UIData{
				DisplayName: "Jira Issue key",
				Hint:        "Jira Issue key under JVS project",
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &JiraPlugin{
				uiData: tc.uiData,
			}

			ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))
			got, err := p.GetUIData(ctx, tc.req)
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreUnexported(jvspb.UIData{})); diff != "" {
				t.Errorf("Failed GetUIData (-want,+got):\n%s", diff)
			}
		})
	}
}
