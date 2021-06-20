# Error Notification

When an error occurs in `lambuild`, `lambuild` sends an error notificaiton to the associated pull request or commit.

For example, if the expression is invalid an error would occur and `lambuild` would fail to start the build.
Then the following comment would be sent to the associated pull request.

![image](https://user-images.githubusercontent.com/13323303/116962766-e7b5c300-ace1-11eb-80ad-70a4291a913c.png)

We can change the message template by the [Lambda Function's configuration](lambda-configuration.md).

If no pull request is associated with the event, the comment is sent to the associated commit.
