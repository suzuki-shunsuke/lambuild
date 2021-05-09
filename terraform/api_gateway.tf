resource "aws_apigatewayv2_api" "main" {
  name          = var.api_gateway_name
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "main" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  payload_format_version = "2.0"

  description        = "lambuild Getting Started"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.main.invoke_arn
}

resource "aws_apigatewayv2_deployment" "main" {
  api_id      = aws_apigatewayv2_route.main.api_id
  description = "lambuild Getting Started"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apigatewayv2_route" "main" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /lambuild"

  target = "integrations/${aws_apigatewayv2_integration.main.id}"
}

resource "aws_lambda_permission" "lambda_permission" {
  statement_id  = "AllowLambuildInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "apigateway.amazonaws.com"

  # The /*/*/* part allows invocation from any stage, method and resource path
  # within API Gateway HTTP API.
  source_arn = "${aws_apigatewayv2_api.main.execution_arn}/*/*/lambuild"
}
