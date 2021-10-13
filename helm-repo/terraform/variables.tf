variable "bucket_name" {
  type        = string
  description = "S3 bucket name for helm repository"
}

variable "rw_access_iam_arns" {
  type        = list(string)
  description = "List of ARNs of IAM entities which will have rw access to bucket"
}

variable "logging_bucket" {
  type        = string
  description = "S3 bucket name for logs"
}

variable "region" {
  type = string
}