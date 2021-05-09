# Secret

`lambuild` requies some secrets.

* GitHub Access Token
  * get pull requests
  * send error notification to related commits or pull requests
* [GitHub Webhook Secret](https://docs.github.com/en/developers/webhooks-and-events/securing-your-webhooks)

We have to store these secrets to AWS Systems Manager Parameter Store.
