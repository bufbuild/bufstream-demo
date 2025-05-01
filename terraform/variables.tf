variable "gcp_project_id" {
  description = "The GCP project ID. Set via TF_VAR_gcp_project_id environment variable."
  type        = string
  # No default value - must be provided
}

variable "gcp_region" {
  description = "The GCP region to deploy resources in (e.g., us-central1). Set via TF_VAR_gcp_region environment variable."
  type        = string
  # No default value - must be provided
}

variable "gcs_bucket_name" {
  description = "The name for the GCS bucket. Set via TF_VAR_gcs_bucket_name environment variable."
  type        = string
  # No default value - must be provided
}

variable "bq_dataset_name" {
  description = "The name for the BigQuery dataset. Set via TF_VAR_bq_dataset_name environment variable."
  type        = string
  # No default value - must be provided
}

variable "bq_connection_id" {
  description = "The simple ID for the BigQuery connection (not the full resource name). Set via TF_VAR_bq_connection_id environment variable."
  type        = string
  # No default value - must be provided
}

variable "bq_location" {
  description = "The location for BigQuery dataset and connection (e.g., US, EU, asia-northeast1). Set via TF_VAR_bq_location environment variable."
  type        = string
  # No default value - must be provided
}

variable "gcp_credentials_file" {
  description = "Optional: Path to the GCP credentials JSON file. Set via TF_VAR_gcp_credentials_file environment variable."
  type        = string
  sensitive   = true
  default     = null
}
