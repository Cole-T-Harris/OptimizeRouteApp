variable "function_name" {
  description = "The name of the Lambda function"
  type        = string
}

variable "handler" {
  description = "The function handler"
  type        = string
}

variable "runtime" {
  description = "The runtime environment for the Lambda function"
  type        = string
}

variable "filename" { 
  description = "Path to the Lambda function deployment package"
  type        = string
}

variable "environment_variables" {
  description = "A map of environment variables for the Lambda function"
  type        = map(string)
  default     = {}
}

variable "permissions" {
  description = "List of permissions for the Lambda function"
  type = list(object({
    principal = string
    source_arn = string
  }))
  default = []
}
