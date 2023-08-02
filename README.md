# Jira Plugin for Justification Verification Service (JVS)

**Jira plugin for JVS is not an official Google product.**

This repository contains components related to the plugin for justification 
verification service.

## Installation

You can use the provided Terraform module to setup the basic infrastructure
needed for this service. Otherwise you can refer to the provided module to see
how to build your own Terraform from scratch.

```terraform
module "jvs" {
  source     = "git::https://github.com/abcxyz/jvs-plugin-jira.git//terraform/example?ref=main" # this should be pinned to the SHA desired
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

  jira_plugin_endpoint = "https://blahblah.atlassian.net/rest/api/3",
  jira_plugin_jql = "project = ABC and assignee != jsmith",
  jira_plugin_account = "jvs-jira-bot@example.com",
  jira_plugin_display_name = "Jira Issue Key",
  jira_plugin_hint = "Jira Issue Key under JVS project",
}
```
