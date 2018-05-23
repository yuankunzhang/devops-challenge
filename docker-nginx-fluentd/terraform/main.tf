provider "aws" {}

resource "aws_s3_bucket" "fluentd_log_bucket" {
  bucket = "devops.challenge.fluentd.log.bucket"
  acl    = "private"

  tags {
    Name        = "Fluentd log bucket"
    Environment = "Dev"
  }
}

data "terraform_remote_state" "store_tf_state" {
  backend = "s3"

  config {
    bucket  = "${aws_s3_bucket.fluentd_log_bucket.bucket}"
    key     = "state_file/terraform.tfstate"
    profile = "default"
  }
}
