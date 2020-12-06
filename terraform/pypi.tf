resource "google_pubsub_topic" "trigger-pypi-topic" {
  name = "trigger-pypi-topic"
}

resource "google_cloud_scheduler_job" "trigger-pypi-scheduler" {
  name        = "trigger-pypi-scheduler"
  description = "The scheduler that triggers fetching new PyPI packages"
  schedule    = "*/5 * * * *"

  pubsub_target {
    topic_name = google_pubsub_topic.trigger-pypi-topic.id
    data       = base64encode(timestamp())
  }
}

resource "google_storage_bucket_object" "function-archive-pypi" {
  name   = "pypi.zip"
  bucket = google_storage_bucket.feed-functions-bucket.name
  source = "../feeds/pypi/pypi.zip"
}

resource "google_cloudfunctions_function" "function-pypi" {
  name        = "feed-pypi-function"
  description = "The Cloud Function that polls PyPI's RSS feeds for new packages to analyze"
  runtime     = "go113"

  available_memory_mb   = 128
  source_archive_bucket = google_storage_bucket.feed-functions-bucket.name
  source_archive_object = google_storage_bucket_object.function-archive-pypi.name

  entry_point = "Poll"

  event_trigger {
    event_type = "google.pubsub.topic.publish"
    resource   = google_pubsub_topic.trigger-pypi-topic.name
  }

  environment_variables = {
    FEED_TOPIC = google_pubsub_topic.feed-topic.name
  }
}
