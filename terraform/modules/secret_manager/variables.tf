variable "project_id" {
  description = "GCP project ID where the secrets belongs."
  type        = string
}

variable "secret_id" {
  description = <<EOT
    ID of the secret. This must be unique within the project. This is the secret
    name.
  EOT
  type        = string
}

variable "labels" {
  description = "The labels assigned to this Secret."
  type        = map(string)
  default     = {}
}

variable "jira_api_token_accessors" {
  description = <<EOT
    List of IAM members with the secret manager secret accessor role. These users will be
    given the role "roles/secretmanager.secretAccessor". Please reference
    https://cloud.google.com/iam/docs/overview#cloud-iam-policy
    for the format for member identifiers.
  EOT
  type        = list(string)
  default     = []
}
