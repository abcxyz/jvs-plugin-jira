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
	"errors"
	"fmt"

	"github.com/abcxyz/pkg/cli"
)

// PluginConfig defines the set over environment variables required
// for running the plugin.
type PluginConfig struct {
	// JIRAEndpoint is the base uri to form the [JIRA REST API uri]. It has the
	// format of:
	//     https://host:port/context/rest/api-name/api-version
	//
	// [JIRA REST API url]: https://developer.atlassian.com/server/jira/platform/rest-apis/#uri-structure
	JIRAEndpoint string

	// Jql is the [JQL] query specifying validation criteria.
	//
	// [JQL]: https://support.atlassian.com/jira-service-management-cloud/docs/use-advanced-search-with-jira-query-language-jql/
	Jql string

	// JIRAAccount is the user name used in [JIRA Basic Auth].
	//
	// [JIRA Basic Auth]: https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
	JIRAAccount string

	// APITokenSecretID is the resource name of the
	// [SecretVersion][google.cloud.secretmanager.v1.SecretVersion] for the API
	// token in the format `projects/*/secrets/*/versions/*`.
	APITokenSecretID string

	// DisplaNname is for display, e.g. for the web UI.
	DisplayName string

	// Hint is for what value to put as the justification.
	Hint string
}

// Validate checks if the config is valid.
func (cfg *PluginConfig) Validate() error {
	var merr error

	if cfg.JIRAEndpoint == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JIRA_PLUGIN_ENDPOINT"))
	}

	if cfg.Jql == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JIRA_PLUGIN_JQL"))
	}

	if cfg.JIRAAccount == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JIRA_PLUGIN_ACCOUNT"))
	}

	if cfg.APITokenSecretID == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JIRA_PLUGIN_API_TOKEN_SECRET_ID"))
	}

	if cfg.Hint == "" {
		merr = errors.Join(merr, fmt.Errorf("empty JIRA_PLUGIN_HINT"))
	}

	return merr
}

// ToFlags binds the config to the give [cli.FlagSet] and returns it.
func (cfg *PluginConfig) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	// Command options
	f := set.NewSection("JIRA PLUGIN OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-endpoint",
		Target:  &cfg.JIRAEndpoint,
		EnvVar:  "JIRA_PLUGIN_ENDPOINT",
		Example: "https://your-domain.atlassian.net/rest/api/3",
		Usage:   "The base uri to form JIRA REST API uri.",
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-jql",
		Target:  &cfg.Jql,
		EnvVar:  "JIRA_PLUGIN_JQL",
		Example: "project = JRA and assignee != jsmith",
		Usage:   "The JQL query specifying validation criteria for a JIRA issue.",
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-account",
		Target:  &cfg.JIRAAccount,
		EnvVar:  "JIRA_PLUGIN_ACCOUNT",
		Example: "abc@xyz.com",
		Usage:   "The user name used in JIRA Basic Auth.",
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-api-token-secret-id",
		Target:  &cfg.APITokenSecretID,
		EnvVar:  "JIRA_PLUGIN_API_TOKEN_SECRET_ID",
		Example: "projects/*/secrets/*/versions/*",
		Usage:   "The resource name of [google.cloud.secretmanager.v1.SecretVersion].",
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-display-name",
		Target:  &cfg.DisplayName,
		EnvVar:  "JIRA_PLUGIN_DISPLAY_NAME",
		Default: "Jira Issue Key",
		Usage:   "The display name is for display, e.g. for the web UI.",
	})

	f.StringVar(&cli.StringVar{
		Name:    "jira-plugin-hint",
		Target:  &cfg.Hint,
		EnvVar:  "JIRA_PLUGIN_HINT",
		Example: "Jira Issue Key under specific project",
		Usage:   "Hint is for what value to put as the justification.",
	})

	return set
}
