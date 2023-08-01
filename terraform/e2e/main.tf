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

  project_id = var.project_id

  # Specify who can access JVS.
  jvs_invoker_members = var.jvs_invoker_members

  # Use your own domain.
  jvs_api_domain = var.jvs_api_domain
  jvs_ui_domain  = var.jvs_ui_domain

  iap_support_email = var.iap_support_email

  jvs_container_image = var.jvs_container_image


  # Specify the Email address where alert notifications send to.
  notification_channel_email = var.notification_channel_email

  jvs_prober_image = var.jvs_prober_image

  # Specify the plugin environment variables. See the file below for details:
  # https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go
  plugin_envvars = {
    "JIRA_PLUGIN_ENDPOINT" : var.jira_plugin_endpoint,
    "JIRA_PLUGIN_JQL" : var.jira_plugin_jql,
    "JIRA_PLUGIN_ACCOUNT" : var.jira_plugin_account,
    # Update to the correct version when the real API token is stored.
    "JIRA_PLUGIN_API_TOKEN_SECRET_ID" : module.jira_api_token.secret_version_name,
    "JIRA_PLUGIN_DISPLAY_NAME" : var.jira_plugin_display_name,
    "JIRA_PLUGIN_HINT" : var.jira_plugin_hint,
  }
}

module "jira_api_token" {
  source = "../modules/secret_manager"

  project_id = var.project_id

  secret_id = "jira_api_token"

  jira_api_token_accessors = [
    "serviceAccount:${module.jvs.jvs_api_service_account_email}",
    "serviceAccount:${module.jvs.jvs_ui_service_account_email}",
  ]
}
