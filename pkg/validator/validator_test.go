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
	"reflect"
	"testing"
)

type fakeJiraEndpoint struct {
	issue       []byte
	matchResult []byte
}

func (srv *fakeJiraEndpoint) Server() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/issue/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", srv.issue)
	})
	mux.HandleFunc("/jql/match", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", srv.matchResult)
	})

	return httptest.NewServer(mux)
}

func TestValidation(t *testing.T) {
	t.Parallel()

	match := Match{
		MatchedIssues: []int{1234},
		Errors:        []string{},
	}
	want := MatchResult{
		Matches: []Match{match},
	}

	jira := fakeJiraEndpoint{
		issue:       []byte(`{"id":"1234","self":"https://test.atlassian.net/rest/api/3/issue/1234","key":"ABCD"}`),
		matchResult: []byte(`{"matches":[{"matchedIssues":[1234],"errors":[]}]}`),
	}
	svr := jira.Server()
	t.Cleanup(func() {
		svr.Close()
	})

	validator, err := NewValidator(svr.URL)
	if err != nil {
		t.Errorf("Cannot create validator: %s", err)
	}
	ctx := context.Background()
	got, err := validator.MatchIssue(ctx, "ABCD", "status NOT IN (Done)", "test@test.com", "secrets")
	if err != nil {
		t.Errorf("Unexpected err: %s", err)
	}
	if !reflect.DeepEqual(want, *got) {
		t.Errorf("Unexpected validation result: want: %v, got: %v", want, got)
	}
}
