# Enable necessary APIs
resource "google_project_service" "bigquery_api" {
  project = var.gcp_project_id
  service = "bigquery.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "bigquery_connection_api" {
  project = var.gcp_project_id
  service = "bigqueryconnection.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "storage_api" {
  project = var.gcp_project_id
  service = "storage.googleapis.com"
  disable_on_destroy = false
}

# GCS Bucket for Iceberg data
resource "google_storage_bucket" "iceberg_bucket" {
  name          = var.gcs_bucket_name
  location      = var.bq_location // Buckets used by BQ connections should be in the same location
  force_destroy = true // Set to false in production

  uniform_bucket_level_access = true

  depends_on = [google_project_service.storage_api]
}

# BigQuery Dataset
resource "google_bigquery_dataset" "iceberg_dataset" {
  dataset_id = var.bq_dataset_name
  project    = var.gcp_project_id
  location   = var.bq_location // Ensure this matches the connection location
  delete_contents_on_destroy = true

  depends_on = [google_project_service.bigquery_api]
}

# BigQuery Connection to GCS
# Note: Using google-beta provider as BigQuery connections might have features in beta
resource "google_bigquery_connection" "gcs_connection" {
  provider      = google-beta
  project       = var.gcp_project_id
  location      = var.bq_location // Must match the dataset location
  connection_id = var.bq_connection_id
  cloud_resource {} // For GCS connection

  depends_on = [google_project_service.bigquery_connection_api]
}

// Grant the Connection Service Account access to the GCS Bucket
resource "google_storage_bucket_iam_member" "connection_bucket_access" {
  bucket = google_storage_bucket.iceberg_bucket.name
  role   = "roles/storage.objectAdmin" // Required for reading/writing data
  member = "serviceAccount:${google_bigquery_connection.gcs_connection.cloud_resource.0.service_account_id}"

  depends_on = [
    google_storage_bucket.iceberg_bucket,
    google_bigquery_connection.gcs_connection
  ]
}

// Output the connection service account ID for verification if needed
output "connection_service_account_id" {
  value = google_bigquery_connection.gcs_connection.cloud_resource.0.service_account_id
  description = "The service account ID associated with the BigQuery GCS connection."
}

output "connection_name_full" {
  value = "projects/${var.gcp_project_id}/locations/${var.bq_location}/connections/${var.bq_connection_id}"
  description = "The full connection name for reference."
}

output "gcs_bucket_name_output" {
  value = google_storage_bucket.iceberg_bucket.name
  description = "The name of the created GCS bucket."
}
