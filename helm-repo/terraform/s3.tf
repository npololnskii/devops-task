
resource "aws_s3_bucket" "helm" {
  bucket = var.bucket_name
  acl    = "private"

  logging {
    content {
      target_bucket = var.logging_bucket
      target_prefix = var.bucket_name
    }
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }

  versioning {
    enabled = true
  }

  tags = {
    Name      = var.bucket_name
    Terraform = true
  }
}

data "aws_iam_policy_document" "rw_policy" {
  dynamic "statement" {
    for_each = var.rw_access_iam_arns

    content {
      sid = "Access for IAM entities"
      actions = [
        "s3:GetObject*",
        "s3:PutObject*"
      ]
      effect    = "Allow"
      resources = ["${aws_s3_bucket.helm.arn}/*"]

      principals {
        identifiers = [statement.value]
        type        = "AWS"
      }
    }
  }
}

resource "aws_s3_bucket_policy" "this" {
  bucket = aws_s3_bucket.helm.id
  policy = data.aws_iam_policy_document.rw_policy.json
}