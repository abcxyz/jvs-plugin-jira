# Copyright 2023 The Authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

module "jvs" {
  source = "git::https://github.com/abcxyz/jvs.git//terraform/e2e?ref=main" # this should be pinned to the SHA desired

  project_id = "YOUR_PROJECT_ID"

  # Specify who can access JVS.
  jvs_invoker_members = ["user:foo@example.com"]

  # Use your own domain.
  jvs_api_domain = "api.jvs.example.com"
  jvs_ui_domain  = "web.jvs.example.com"

  iap_support_email = "support@example.com"

  jvs_container_image = "us-docker.pkg.dev/abcxyz-artifacts/docker-images/jvsctl:0.0.5-amd64"


  # Specify the Email address where alert notifications send to.
  notification_channel_email = "foo@example.com"

  jvs_prober_image = "us-docker.pkg.dev/abcxyz-artifacts/docker-images/jvs-prober:0.0.5-amd64"

  # Specify the plugin environment variables. See the file below for details:
  # https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go
  plugin_envvars = {
    "JIRA_PLUGIN_ENDPOINT" : "https://blahblah.atlassian.net/rest/api/3",
    "JIRA_PLUGIN_JQL" : "project = ABC and assignee != jsmith",
    "JIRA_PLUGIN_ACCOUNT" : "jvs-jira-bot@example.com",
    # Update to the correct version when the real API token is stored.
    "JIRA_PLUGIN_API_TOKEN_SECRET_ID" : module.jira_api_token.secret_version_name,
    "JIRA_PLUGIN_DISPLAY_NAME" : "Jira Issue Key",
    "JIRA_PLUGIN_HINT" : "Jira Issue Key under JVS project",
  }
}

module "jira_api_token" {
  # Pin to proper version.
  source = "git::https://github.com/abcxyz/jvs-plugin-jira.git//terraform/modules/secret_manager?ref=5c0eb018b293c3be3cf0e6a572f190e66c6e9f66"

  project_id = "YOUR_PROJECT_ID"

  secret_id = "jira_api_token"
}

resource "google_secret_manager_secret_iam_member" "jira_api_token_accessor" {
  for_each = {
    api_sa = "serviceAccount:${module.jvs.jvs_api_service_account_email}"
    ui_sa  = "serviceAccount:${module.jvs.jvs_ui_service_account_email}"
  }

  project = "YOUR_PROJECT_ID"

  secret_id = module.jira_api_token.secret_id

  role = "roles/secretmanager.secretAccessor"

  member = each.value
}

resource "google_secret_manager_secret_iam_member" "jira_api_token_admin" {
  project = "YOUR_PROJECT_ID"

  secret_id = module.jira_api_token.secret_id

  role = "roles/secretmanager.admin"

  member = "group:jira-api-token-admins@example.com"
}
