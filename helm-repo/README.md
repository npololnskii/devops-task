### Helm s3 repo

Contains terraform code to spin up secure s3 bucket for helm repository. 

Bucket has versioning, server side encryption and logging enbled. 

You can provide list of IAM entities as rw_access_iam_arns var which will have RW permissions for bucket entries. 


Run terraform plan and apply to create bucket.

```bash
terraform plan -out "start"
terraform apply "start"
```

```bash
# Install helm s3 plugin
helm3 plugin install https://github.com/hypnoglow/helm-s3.git

# Init repo
helm3 s3 init s3://helm-tests/charts

```