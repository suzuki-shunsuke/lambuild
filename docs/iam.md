# Lambda Execution Role

## read-secret

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ssm:GetParameter",
            "Resource": [
                "arn:aws:ssm:<region>:<aws account id>:parameter/<SSM_PARAMETER_NAME_GITHUB_TOKEN >",
                "arn:aws:ssm:<region>:<aws account id>:parameter/<SSM_PARAMETER_NAME_WEBHOOK_SECRET >"
            ]
        }
    ]
}
```

## start-codebuild

Please restrict `Resource` as you like.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "codebuild:StartBuildBatch",
                "codebuild:StartBuild"
            ],
            "Resource": "*"
        }
    ]
}
```
