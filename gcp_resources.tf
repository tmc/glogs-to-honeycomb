# Sample terraform resources.

# Service account for glogs-to-honeycomb service.
resource "google_service_account" "glogs-to-honeycomb" {
  project      = local.project
  account_id   = "svc-glogs-to-honeycomb"
  display_name = "Service account for glogs-to-honeycomb"
  description  = "Service account for glogs-to-honeycomb, a pubsub subscriber that ships logs to honeycomb."
}

# PubSub topic for log drain entries relating to Istio sidecar logs.
resource "google_pubsub_topic" "istio-sidecar-log-sink" {
  project = local.project
  name    = "istio-sidecar-log-sink"
}

# Subscription for istio-sidecar-log-sink topic.
resource "google_pubsub_subscription" "istio-sidecar-log-sink" {
  project               = local.project
  name                  = "istio-sidecar-log-sink"
  topic                 = google_pubsub_topic.istio-sidecar-log-sink.name
  ack_deadline_seconds  = 10
  retain_acked_messages = true
  retry_policy {
    minimum_backoff = "1s"
    maximum_backoff = "60s"
  }
}

# Ship Istio sidecar logs to a particular pubsub bucket.
resource "google_logging_project_sink" "istio-sidecar-log-sink" {
  project     = local.project
  name        = "istio-sidecar-log-sink"
  destination = "pubsub.googleapis.com/${google_pubsub_topic.istio-sidecar-log-sink.id}"
  filter      = <<EOT
resource.type="k8s_container"
resource.labels.container_name="istio-proxy"
-resource.labels.namespace_name="virta-system"
EOT
}

# Grants publish access to the built-in gcloud service account for log shipping to pubsub.
resource "google_pubsub_topic_iam_member" "istio-sidecar-log-sink" {
  project = local.project
  topic   = google_pubsub_topic.istio-sidecar-log-sink.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:cloud-logs@system.gserviceaccount.com"
}

# Grants glogs-to-honeycomb access to istio logs subscription.
resource "google_pubsub_subscription_iam_member" "svc-glogs-to-honeycomb-istio-logs" {
  project      = local.project
  subscription = google_pubsub_subscription.istio-sidecar-log-sink.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${google_service_account.glogs-to-honeycomb.email}"
}
