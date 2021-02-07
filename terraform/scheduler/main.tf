resource "google_cloud_scheduler_job" "trigger-ecosystem-scheduler" {
  name        = "trigger-feeds-scheduler"
  description = "Scheduler for packages from feeds"
  schedule    = "*/5 * * * *"

  http_target {
    http_method = "POST"
    uri         = google_cloud_run_service.run-scheduler.status[0].url

    oidc_token {
      service_account_email = var.service-account-email
    }
  }
}

resource "google_cloud_run_service" "run-scheduler" {
  name     = "scheduled-feeds-srv"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project}/scheduled-feeds"
        env {
          name  = "OSSMALWARE_TOPIC_URL"
          value = var.pubsub-topic-feed-id
        }
      }
    }
  }
}
