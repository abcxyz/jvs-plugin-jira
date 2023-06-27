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
	"errors"
	"fmt"

	"github.com/abcxyz/pkg/cli"
)

// PluginConfig defines the set over environment variables required
// for running the plugin.
type PluginConfig struct {
	// JiraEndpoint is the base uri to form the [JIRA REST API uri]. It has the
	// format of:
	//     http://host:port/context/rest/api-name/api-version
	//
	// [JIRA REST API url]: https://developer.atlassian.com/server/jira/platform/rest-apis/#uri-structure
	JiraEndpoint string `env:"JIRA_PLUGIN_ENDPOINT"`

	// Jql is the [JQL] query specifying validation criteria.
	//
	// [JQL]: https://support.atlassian.com/jira-service-management-cloud/docs/use-advanced-search-with-jira-query-language-jql/
	Jql string `env:"JIRA_PLUGIN_JQL"`

	// JiraAccount is the user name used in [JIRA Basic Auth].
	//
	// [JIRA Basic Auth]: https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
	JiraAccount string `env:"JIRA_PLUGIN_ACCOUNT"`

	// ApiTokenSecretID is the resource name of the
	// [SecretVersion][google.cloud.secretmanager.v1.SecretVersion] for the API
	// token in the format `projects/*/secrets/*/versions/*`.
	ApiTokenSecretID string `env:"JIRA_PLUGIN_API_TOKEN_SECRET_ID"`
}

// Validate checks if the config is valid.
func (cfg *PluginConfig) Validate() error {
	var merr error

	if cfg.JiraEndpoint == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JiraEndpoint"))
	}

	if cfg.Jql == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JQL"))
	}

	if cfg.JiraAccount == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JiraAccount"))
	}

	if cfg.ApiTokenSecretID == "" {
		merr = errors.Join(merr, fmt.Errorf("empty ApiTokenSecretID"))
	}

	return merr
}

// ToFlags binds the config to the give [cli.FlagSet] and returns it.
func (cfg *PluginConfig) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	// Command options
	f := set.NewSection("JIRA PLUGIN OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-endpoint",
		Target:  &cfg.JiraEndpoint,
		EnvVar:  "JIRA_PLUGIN_ENDPOINT",
		Example: "https://your-domain.atlassian.net/rest/api/3",
		Usage: `
		JiraEndpoint is the base uri to form the [JIRA REST API uri]. It has the
		format of:
		    http://host:port/context/rest/api-name/api-version
		
		[JIRA REST API url]: https://developer.atlassian.com/server/jira/platform/rest-apis/#uri-structure
		`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-jql",
		Target:  &cfg.Jql,
		EnvVar:  "JIRA_PLUGIN_JQL",
		Example: "project = JRA and assignee != jsmith",
		Usage: `
		Jql is the [JQL] query specifying validation criteria.
		
		[JQL]: https://support.atlassian.com/jira-service-management-cloud/docs/use-advanced-search-with-jira-query-language-jql/
		`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-account",
		Target:  &cfg.JiraAccount,
		EnvVar:  "JIRA_PLUGIN_ACCOUNT",
		Example: "abc@xyz.com",
		Usage: `
		JiraAccount is the user name used in [JIRA Basic Auth].
		
		[JIRA Basic Auth]: https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
		`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-api-token-secret-id",
		Target:  &cfg.ApiTokenSecretID,
		EnvVar:  "JIRA_PLUGIN_API_TOKEN_SECRET_ID",
		Example: "projects/*/secrets/*/versions/*",
		Usage: `
		ApiTokenSecretID is the resource name of the
		[SecretVersion][google.cloud.secretmanager.v1.SecretVersion] for the API
		token in the format "projects/*/secrets/*/versions/*".
		`,
	})

	return set
}
