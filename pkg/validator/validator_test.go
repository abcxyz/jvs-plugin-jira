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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		issuesHandler http.Handler
		matchHandler  http.Handler
		want          *MatchResult
		wantErr       string
	}{
		{
			name: "happy_path",
			issuesHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"id":"1234","self":"https://test.atlassian.net/rest/api/3/issue/1234","key":"ABCD"}`)
			}),
			matchHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"matches":[{"matchedIssues":[1234],"errors":[]}]}`)
			}),
			want: &MatchResult{
				Matches: []*Match{
					{
						MatchedIssues: []int{1234},
						Errors:        []string{},
					},
				},
			},
		},
		{
			name: "none_match_jql",
			issuesHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"id":"1234","self":"https://test.atlassian.net/rest/api/3/issue/1234","key":"ABCD"}`)
			}),
			matchHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"matches":[{"matchedIssues":[],"errors":[]}]}`)
			}),
			want: &MatchResult{
				Matches: []*Match{
					{
						MatchedIssues: []int{},
						Errors:        []string{},
					},
				},
			},
		},
		{
			name: "invalid_match_request",
			issuesHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"id":"1234","self":"https://test.atlassian.net/rest/api/3/issue/1234","key":"ABCD"}`)
			}),
			matchHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, `{"errorMessages":["There was an error parsing JSON. Check that your request body is valid."]}`)
			}),
			want:    nil,
			wantErr: "response code 400",
		},
		{
			name: "issue_does_not_exist",
			issuesHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, `{"errorMessages":["Issue does not exist or you do not have permission to see it."],"errors":{}}`)
			}),
			matchHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"matches":[{"matchedIssues":[1234],"errors":[]}]}`)
			}),
			want:    nil,
			wantErr: "response code 404",
		},
		// TODO: add failure test cases.
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			mux.Handle("/issue/", tc.issuesHandler)
			mux.Handle("/jql/match/", tc.matchHandler)

			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			validator, err := NewValidator(srv.URL, "status NOT IN (Done)", "test@test.com", "secrets")
			if err != nil {
				t.Fatalf("failed to create validator: %v", err)
			}

			ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))
			got, err := validator.MatchIssue(ctx, "ABCD")
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Failed validation (-want,+got):\n%s", diff)
			}
		})
	}
}
