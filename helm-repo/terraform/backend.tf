provider "aws" {
  region = var.region
}

// Set local backend but better to use s3 backend 
terraform {
  backend "local" {
  }
}
