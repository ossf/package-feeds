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
  autogenerate_revision_name = true

  metadata {
    annotations = {
      "autoscaling.knative.dev/maxScale" = "1"
      "autoscaling.knative.dev/minScale" = "1"
    }
  }
  template {
    spec {
      container_concurrency = 1
      containers {
        image = "gcr.io/${var.project}/scheduled-feeds"
        env {
          name  = "OSSMALWARE_TOPIC_URL"
          value = "gcppubsub://${var.pubsub-topic-feed-id}"
        }
        resources {
          limits = {
            memory = "2Gi"
            cpu = "1000m"
          }
        }
      }
    }
  }
}
