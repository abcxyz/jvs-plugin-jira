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

variable "project_id" {
  description = "The GCP project to host the justification verification service."
  type        = string
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "The default Google Cloud region to deploy resources in (defaults to 'us-central1')."
}

variable "jvs_invoker_members" {
  description = "The list of members that can call JVS."
  type        = list(string)
  default     = []
}

variable "jvs_api_domain" {
  description = "The JVS API domain."
  type        = string
}

variable "jvs_ui_domain" {
  description = "The JVS UI domain."
  type        = string
}

variable "iap_support_email" {
  description = "The IAP support email."
  type        = string
}

variable "kms_key_location" {
  type        = string
  default     = "global"
  description = "The location where kms key will be created."
}

variable "jvs_container_image" {
  description = "Container image for the jvsctl CLI and server entrypoints."
  type        = string
}

variable "notification_channel_email" {
  type        = string
  description = "The Email address where alert notifications send to."
}

variable "jvs_prober_image" {
  type        = string
  description = "docker image for jvs-prober"
}

variable "jira_plugin_endpoint" {
  type        = string
  description = <<EOT
      Base uri to form JIRA REST API uri. Please see 
      https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go 
      for more details.
    EOT
}

variable "jira_plugin_jql" {
  type        = string
  description = <<EOT
      JQL to specify jira validation criteria. Please see 
      https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go 
      for more details.
    EOT
}

variable "jira_plugin_account" {
  type        = string
  description = <<EOT
      JIRA account to communicate to JIRA REST API. Please see 
      https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go 
      for more details.
    EOT
}

variable "jira_plugin_display_name" {
  type        = string
  description = <<EOT
      Category text to display in the select drop down. Please see 
      https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go 
      for more details.
    EOT
}

variable "jira_plugin_hint" {
  type        = string
  description = <<EOT
      Ghost hint to display when jira category is selected. Please see 
      https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go 
      for more details.
    EOT
}
