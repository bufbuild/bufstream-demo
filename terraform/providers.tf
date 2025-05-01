terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project     = var.gcp_project_id
  region      = var.gcp_region
  credentials = var.gcp_credentials_file != null ? file(var.gcp_credentials_file) : null
}

provider "google-beta" {
  project     = var.gcp_project_id
  region      = var.gcp_region
  credentials = var.gcp_credentials_file != null ? file(var.gcp_credentials_file) : null
}
