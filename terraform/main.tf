provider "aws" {
  region = "us-east-1"
}

module "optimize_route_function" {
  source            = "./lambda_module"
  function_name     = "optimize_route_function"
  handler           = "handler1"
  runtime           = "provided.al2023"
  filename          = "../dist/optimizeRoute/optimizeRoute.zip"
  environment_variables = {
    GOOGLE_API_KEY: var.GOOGLE_API_KEY
    SUPABASE_URL: var.SUPABASE_URL
    SUPABASE_KEY: var.SUPABASE_KEY  
  }
  lambda_timeout    = 15
}

module "commutes_queue_function" {
  source            = "./lambda_module"
  function_name     = "commutes_queue_function"
  handler           = "handler1"
  runtime           = "provided.al2023"
  filename          = "../dist/commutesQueue/commutesQueue.zip"
  environment_variables = {
    SUPABASE_USERNAME: var.SUPABASE_USERNAME
    SUPABASE_PASSWORD: var.SUPABASE_PASSWORD
    SUPABASE_HOST: var.SUPABASE_HOST
    SUPABASE_PORT: var.SUPABASE_PORT
    SUPABASE_DATABASE: var.SUPABASE_DATABASE
    OPTIMIZE_ROUTE_FUNCTION: module.optimize_route_function.function_arn
  }
    lambda_timeout    = 180
}
