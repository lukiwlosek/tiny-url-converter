# Input variable definitions

variable "aws_region" {
  description = "AWS region for all resources."

  type    = string
  default = "us-east-1"
}

variable "profile" {
  description = "AWS profile used"

  type  = string
  default = "cloudsoft_cf_test"
}
