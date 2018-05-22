provider "aws" {}

resource "aws_s3_bucket" "fluentd_log_bucket" {
  bucket = "devops.challenge.fluentd.log.bucket"
  acl    = "private"

  tags {
    Name        = "Fluentd log bucket"
    Environment = "Dev"
  }
}
