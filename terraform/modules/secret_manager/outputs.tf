output "secret_id" {
  value = google_secret_manager_secret.jira_api_token.id
}

output "secret_name" {
  value = google_secret_manager_secret.jira_api_token.name
}
