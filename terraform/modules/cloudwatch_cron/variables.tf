variable "rule_name" {
  description = "The name of the CloudWatch Events rule"
  type        = string
}

variable "rule_description" {
  description = "The description of the CloudWatch Events rule"
  type        = string
  default     = ""
}

variable "schedule_expression" {
  description = "The schedule expression for the CloudWatch Events rule"
  type        = string
}

variable "lambda_function_arn" {
  description = "The ARN of the Lambda function to be triggered"
  type        = string
}

variable "lambda_function_name" {
  description = "The name of the Lambda function to be triggered"
  type        = string
}
