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
				"JIRA_PLUGIN_DISPLAY_NAME":        "Jira Issue Key",
				"JIRA_PLUGIN_HINT":                "Jira Issue Key under JVS project",
				"JIRA_PLUGIN_BASE_URL":            "https://verily-okta-sandbox.atlassian.net/browse/",
			},
			wantConfig: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
				DisplayName:      "Jira Issue Key",
				Hint:             "Jira Issue Key under JVS project",
				BaseURL:          "https://verily-okta-sandbox.atlassian.net/browse/",
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
				"-jira-plugin-display-name", "Jira Issue Key",
				"-jira-plugin-hint", "Jira Issue Key under specific project",
				"-jira-plugin-base-url", "https://verily-okta-sandbox.atlassian.net/browse/",
			},
			wantConfig: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
				DisplayName:      "Jira Issue Key",
				Hint:             "Jira Issue Key under specific project",
				BaseURL:          "https://verily-okta-sandbox.atlassian.net/browse/",
			},
		},
		{
			name: "endpoint_env",
			envs: map[string]string{
				"JIRA_PLUGIN_ENDPOINT": "https://blahblah.atlassian.net/rest/api/3",
			},
			wantConfig: &PluginConfig{
				JIRAEndpoint: "https://blahblah.atlassian.net/rest/api/3",
				DisplayName:  "Jira Issue Key",
			},
		},
		{
			name: "endpoint_flag",
			args: []string{"-jira-plugin-endpoint", "https://blahblah.atlassian.net/rest/api/3"},
			wantConfig: &PluginConfig{
				JIRAEndpoint: "https://blahblah.atlassian.net/rest/api/3",
				DisplayName:  "Jira Issue Key",
			},
		},
		{
			name: "jql_env",
			envs: map[string]string{
				"JIRA_PLUGIN_JQL": "project = JRA and assignee != jsmith",
			},
			wantConfig: &PluginConfig{
				Jql:         "project = JRA and assignee != jsmith",
				DisplayName: "Jira Issue Key",
			},
		},
		{
			name: "jql_flag",
			args: []string{"-jira-plugin-jql", "project = JRA and assignee != jsmith"},
			wantConfig: &PluginConfig{
				Jql:         "project = JRA and assignee != jsmith",
				DisplayName: "Jira Issue Key",
			},
		},
		{
			name: "account_env",
			envs: map[string]string{
				"JIRA_PLUGIN_ACCOUNT": "abc@xyz.com",
			},
			wantConfig: &PluginConfig{
				JIRAAccount: "abc@xyz.com",
				DisplayName: "Jira Issue Key",
			},
		},
		{
			name: "account_flag",
			args: []string{"-jira-plugin-account", "abc@xyz.com"},
			wantConfig: &PluginConfig{
				JIRAAccount: "abc@xyz.com",
				DisplayName: "Jira Issue Key",
			},
		},
		{
			name: "api_token_secret_id_env",
			envs: map[string]string{
				"JIRA_PLUGIN_API_TOKEN_SECRET_ID": "projects/123456/secrets/api-token/versions/4",
			},
			wantConfig: &PluginConfig{
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
				DisplayName:      "Jira Issue Key",
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
				DisplayName:      "Jira Issue Key",
			},
		},
		{
			name: "display_name_env",
			envs: map[string]string{
				"JIRA_PLUGIN_DISPLAY_NAME": "jira display name",
			},
			wantConfig: &PluginConfig{
				DisplayName: "jira display name",
			},
		},
		{
			name: "display_name_flag",
			args: []string{
				"-jira-plugin-display-name", "jira display name",
			},
			wantConfig: &PluginConfig{
				DisplayName: "jira display name",
			},
		},
		{
			name: "hint_env",
			envs: map[string]string{
				"JIRA_PLUGIN_HINT": "jira hint",
			},
			wantConfig: &PluginConfig{
				DisplayName: "Jira Issue Key",
				Hint:        "jira hint",
			},
		},
		{
			name: "hint_flag",
			args: []string{
				"-jira-plugin-hint", "jira hint",
			},
			wantConfig: &PluginConfig{
				DisplayName: "Jira Issue Key",
				Hint:        "jira hint",
			},
		},
		{
			name: "base_url_env",
			envs: map[string]string{
				"JIRA_PLUGIN_BASE_URL": "https://verily-okta-sandbox.atlassian.net/browse/",
			},
			wantConfig: &PluginConfig{
				DisplayName: "Jira Issue Key",
				BaseURL:     "https://verily-okta-sandbox.atlassian.net/browse/",
			},
		},
		{
			name: "base_url_flag",
			args: []string{
				"-jira-plugin-base-url", "https://verily-okta-sandbox.atlassian.net/browse/",
			},
			wantConfig: &PluginConfig{
				DisplayName: "Jira Issue Key",
				BaseURL:     "https://verily-okta-sandbox.atlassian.net/browse/",
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
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
				DisplayName:      "Jira Issue Key",
				Hint:             "Jira Issue Key under JVS project",
				BaseURL:          "https://verily-okta-sandbox.atlassian.net/browse/",
			},
		},
		{
			name: "valid_without_display_name",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
				Hint:             "Jira Issue Key under JVS project",
				BaseURL:          "https://verily-okta-sandbox.atlassian.net/browse/",
			},
		},
		{
			name: "empty_jira_endpoint",
			cfg: &PluginConfig{
				Jql:              "project = JRA and assignee != jsmith",
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
			wantErr: "empty JIRA_PLUGIN_ENDPOINT",
		},
		{
			name: "empty_jql",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
			wantErr: "empty JIRA_PLUGIN_JQL",
		},
		{
			name: "empty_jira_account",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
			wantErr: "empty JIRA_PLUGIN_ACCOUNT",
		},
		{
			name: "empty_api_token_secret_id",
			cfg: &PluginConfig{
				JIRAEndpoint: "https://blahblah.atlassian.net/rest/api/3",
				Jql:          "project = JRA and assignee != jsmith",
				JIRAAccount:  "abc@xyz.com",
			},
			wantErr: "empty JIRA_PLUGIN_API_TOKEN_SECRET_ID",
		},
		{
			name: "empty_hint",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
			},
			wantErr: "empty JIRA_PLUGIN_HINT",
		},
		{
			name: "empty_base_url",
			cfg: &PluginConfig{
				JIRAEndpoint:     "https://blahblah.atlassian.net/rest/api/3",
				Jql:              "project = JRA and assignee != jsmith",
				JIRAAccount:      "abc@xyz.com",
				APITokenSecretID: "projects/123456/secrets/api-token/versions/4",
				Hint:             "Jira Issue Key under JVS project",
			},
			wantErr: "empty JIRA_PLUGIN_BASE_URL",
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
