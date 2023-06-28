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
	"context"
	"fmt"
	"testing"

	"github.com/abcxyz/jvs-plugin-jira/pkg/validator"
	jvspb "github.com/abcxyz/jvs/apis/v0"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type mockValidator struct {
	result *validator.MatchResult
	err    error
}

func (m *mockValidator) MatchIssue(ctx context.Context, issueKey string) (*validator.MatchResult, error) {
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
				result: &validator.MatchResult{
					Matches: []*validator.Match{
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
				result: &validator.MatchResult{
					Matches: []*validator.Match{
						{
							MatchedIssues: []int{1234},
							Errors:        []string{},
						},
					},
				},
			},
			want:    nil,
			wantErr: "failed to perform validation, expected category \"github\" to be \"jira\"",
		},
		{
			name: "empty_value",
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "jira",
				},
			},
			validator: &mockValidator{
				result: &validator.MatchResult{
					Matches: []*validator.Match{
						{
							MatchedIssues: []int{1234},
							Errors:        []string{},
						},
					},
				},
			},
			want:    nil,
			wantErr: "empty justification value",
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
				result: &validator.MatchResult{
					Matches: []*validator.Match{
						{
							MatchedIssues: []int{},
							Errors:        []string{"not match"},
						},
					},
				},
			},
			want: &jvspb.ValidateJustificationResponse{
				Valid:   false,
				Warning: []string{"not match"},
			},
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
				err: fmt.Errorf("failed match"),
			},
			want: &jvspb.ValidateJustificationResponse{
				Valid: false,
				Error: []string{"failed match"},
			},
			wantErr: "failed to validate justification",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &JiraPlugin{
				validator: tc.validator,
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
