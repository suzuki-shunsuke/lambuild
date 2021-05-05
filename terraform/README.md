# Getting Started with Terraform

We provide Terraform Configuration to setup AWS resources quickly for trial.
Please see [Getting Started](/docs/getting-started.md) too.
_Note that this Terraform Configuration is for not production use but getting started._

## Requirement

* Terraform
* AWS Access Key

## Procedure

If you use [tfenv](https://github.com/tfutils/tfenv),

```console
$ tfenv install
```

[Download a zip file from Release page](https://github.com/suzuki-shunsuke/lambuild/releases) on this directory.

Create `config.yaml` from the template.

```console
$ cp config.yaml.template config.yaml
$ vi config.yaml
```

Configure Terraform [input variables](https://www.terraform.io/docs/language/values/variables.html).

```console
$ cp terraform.tfvars.template terraform.tfvars
$ vi terraform.tfvars
```

Create resources.

```console
$ terraform apply [-refresh=false]
```

`-refresh=false` is useful to make terraform commands fast.

## Clean up

```
$ terraform destroy
```
