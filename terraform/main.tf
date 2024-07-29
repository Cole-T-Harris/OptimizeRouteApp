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
}
