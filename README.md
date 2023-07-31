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
  source     = "git::https://github.com/abcxyz/jvs.git//terraform/e2e?ref=main" # this should be pinned to the SHA desired
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

  # The cloud app id for jvs api service.
  prober_audience = "https://example-api-jvs-cloud-run.run.app"

  # Specify the plugin environment variables. See the file below for details:
  # https://github.com/abcxyz/jvs-plugin-jira/blob/main/pkg/plugin/config.go
  plugin_envvars = {
    "JIRA_PLUGIN_ENDPOINT" : "https://blahblah.atlassian.net/rest/api/3",
    "JIRA_PLUGIN_JQL" : "project = JRA and assignee != jsmith",
    "JIRA_PLUGIN_ACCOUNT" : "abc@xyz.com",
    "JIRA_PLUGIN_API_TOKEN_SECRET_ID" : "projects/123456/secrets/api-token/versions/4",
    "JIRA_PLUGIN_DISPLAY_NAME" : "Jira Issue Key",
    "JIRA_PLUGIN_HINT" : "Jira Issue Key under JVS project",
  }
}
```
