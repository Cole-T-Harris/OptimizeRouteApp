resource "aws_lambda_function" "this" {
  function_name = var.function_name
  role          = aws_iam_role.lambda_role.arn
  handler       = var.handler
  runtime       = var.runtime
  filename      = var.filename

  environment {
    variables = var.environment_variables
  }

  timeout = var.lambda_timeout
}

resource "aws_lambda_permission" "this" {
  count = length(var.permissions)

  statement_id  = "AllowInvoke-${var.permissions[count.index].principal}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this.function_name
  principal     = var.permissions[count.index].principal
  source_arn    = var.permissions[count.index].source_arn
}
