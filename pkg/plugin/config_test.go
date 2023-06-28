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
	"testing"

	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestPluginConfig_ToFlags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		args       []string
		envs       map[string]string
		wantConfig *PluginConfig
	}{
		{
			name: "all_envs_specified",
			envs: map[string]string{
				"JIRA_PLUGIN_ENDPOINT":            "https://blahblah.atlassian.net/rest/api/3",
				"JIRA_PLUGIN_JQL":                 "project = JRA and assignee != jsmith",
				"JIRA_PLUGIN_ACCOUNT":             "abc@xyz.com",
				"JIRA_PLUGIN_API_TOKEN_SECRET_ID": "projects/123456/secrets/api-token/versions/4",
			},
			wantConfig: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JiraAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
		},
		{
			name: "all_flags_specified",
			args: []string{
				"-jira-plugin-endpoint", "https://blahblah.atlassian.net/rest/api/3",
				"-jira-plugin-jql", "project = JRA and assignee != jsmith",
				"-jira-plugin-account", "abc@xyz.com",
				"-jira-plugin-api-token-secret-id",
				"projects/123456/secrets/api-token/versions/4",
			},
			wantConfig: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JiraAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
		},
		{
			name: "endpoint_env",
			envs: map[string]string{
				"JIRA_PLUGIN_ENDPOINT": "https://blahblah.atlassian.net/rest/api/3",
			},
			wantConfig: &PluginConfig{
				JIRAEndpoint: "https://blahblah.atlassian.net/rest/api/3",
			},
		},
		{
			name: "endpoint_flag",
			args: []string{"-jira-plugin-endpoint", "https://blahblah.atlassian.net/rest/api/3"},
			wantConfig: &PluginConfig{
				JIRAEndpoint: "https://blahblah.atlassian.net/rest/api/3",
			},
		},
		{
			name: "jql_env",
			envs: map[string]string{
				"JIRA_PLUGIN_JQL": "project = JRA and assignee != jsmith",
			},
			wantConfig: &PluginConfig{
				Jql: "project = JRA and assignee != jsmith",
			},
		},
		{
			name: "jql_flag",
			args: []string{"-jira-plugin-jql", "project = JRA and assignee != jsmith"},
			wantConfig: &PluginConfig{
				Jql: "project = JRA and assignee != jsmith",
			},
		},
		{
			name: "account_env",
			envs: map[string]string{
				"JIRA_PLUGIN_ACCOUNT": "abc@xyz.com",
			},
			wantConfig: &PluginConfig{
				JiraAccount: "abc@xyz.com",
			},
		},
		{
			name: "account_flag",
			args: []string{"-jira-plugin-account", "abc@xyz.com"},
			wantConfig: &PluginConfig{
				JiraAccount: "abc@xyz.com",
			},
		},
		{
			name: "api_token_secret_id_env",
			envs: map[string]string{
				"JIRA_PLUGIN_API_TOKEN_SECRET_ID": "projects/123456/secrets/api-token/versions/4",
			},
			wantConfig: &PluginConfig{
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
		},
		{
			name: "api_token_secret_id_flag",
			args: []string{
				"-jira-plugin-api-token-secret-id",
				"projects/123456/secrets/api-token/versions/4",
			},
			wantConfig: &PluginConfig{
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotConfig := &PluginConfig{}
			set := cli.NewFlagSet(cli.WithLookupEnv(cli.MapLookuper(tc.envs)))
			set = gotConfig.ToFlags(set)
			if err := set.Parse(tc.args); err != nil {
				t.Errorf("unexpected flag set parse error: %v", err)
			}
			if diff := cmp.Diff(tc.wantConfig, gotConfig); diff != "" {
				t.Errorf("Config unexpected diff (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestPluginConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cfg     *PluginConfig
		wantErr string
	}{
		{
			name: "valid",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JiraAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
		},
		{
			name: "empty_jira_endpoint",
			cfg: &PluginConfig{
				Jql:              "project = JRA and assignee != jsmith",
				JiraAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
			wantErr: "empty JIRAEndpoint",
		},
		{
			name: "empty_jql",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				JiraAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
			wantErr: "empty JQL",
		},
		{
			name: "empty_jira_account",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
			wantErr: "empty JiraAccount",
		},
		{
			name: "empty_api_token_secret_id",
			cfg: &PluginConfig{
				JIRAEndpoint: "https://blahblah.atlassian.net/rest/api/3",
				Jql:          "project = JRA and assignee != jsmith",
				JiraAccount:  "abc@xyz.com",
			},
			wantErr: "empty APITokenSecretID",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Unexpected err: %s", diff)
			}
		})
	}
}
