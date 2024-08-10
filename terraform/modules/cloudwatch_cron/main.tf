resource "aws_cloudwatch_event_rule" "this" {
  name                 = var.rule_name
  description          = var.rule_description
  schedule_expression  = var.schedule_expression
}

resource "aws_cloudwatch_event_target" "this" {
  rule      = aws_cloudwatch_event_rule.this.name
  arn       = var.lambda_function_arn
  target_id = "lambda_target"
}

resource "aws_lambda_permission" "this" {
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "events.amazonaws.com"
  statement_id  = "AllowExecutionFromEvents"
  source_arn    = aws_cloudwatch_event_rule.this.arn
}
