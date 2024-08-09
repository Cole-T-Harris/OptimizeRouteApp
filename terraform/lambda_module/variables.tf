variable "environment_variables" {
  description = "A map of environment variables for the Lambda function"
  type        = map(string)
  default     = {}
}

variable "filename" { 
  description = "Path to the Lambda function deployment package"
  type        = string
}

variable "function_name" {
  description = "The name of the Lambda function"
  type        = string
}

variable "handler" {
  description = "The function handler"
  type        = string
}

variable "lambda_timeout" {
  description = "The amount of time that Lambda allows a function to run before stopping it"
  type        = number
  default     = 60  # Set a default value, e.g., 60 seconds
}

variable "permissions" {
  description = "List of permissions for the Lambda function"
  type = list(object({
    principal = string
    source_arn = string
  }))
  default = []
}

variable "runtime" {
  description = "The runtime environment for the Lambda function"
  type        = string
}

