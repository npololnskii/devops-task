# Simple s3 web client 
### Supported endpoints:
```http
GET /files?nextToken=some_token
```
| Parameter | Type | Description |
| :--- | :--- | :--- |
| `nextToken` | `string` | **Optional**. The next /files requests can be continued with this nextToken. Helps to fetch next subset of data.|

#### Response 
```javascript
{
  "files" : [{
    "key" : string,
    "storage_classs": string,
    "last_modified": string,
    "size": int
  }],
  "nextToken" : string
}
```

```http
GET /metrics
```
### Response 

List of the metrics compatible for Prometheus scraping

```http
GET /health
```
### Response 

Return current date and shows that app is running 

```http
GET /isready
```
### Response 

Checks s3 api availabilty and returns if app is ready for request proccessing

## Build 
Assume that golang is already installed on your PC. 
```bash
go get -d -v ./...
go build -o app
```

To build docker image run 
```bash
docker build -t s3-explorer .
```

## Deployment 
This tool can be deployed into k8s cluster with helm chart. 

Some helm chart values:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image | object |  | Defines image name tag and pullPolicy |
| configuration.args |  string |  | Args for container. The same as args for s3-explorer app |
| configuration.env.normal | object | `` | Object with  container's env vars  |
| configuration.env.secret | object | `` | Object with container's secrets mounted as env vars. Also will be provisioned k8s secret with these values.  |
| ingress | object |  |Ingress resource configuration |
| livenessProbe | object |  |livenessProbe configuration|
| readinessProbe | object |  |readinessProbe configuration|

Prepare custom values file:
``` yaml
image:
  repository: s3-explorer
configuration:
  args:
    - "--bucket=your-bucket-name"
    - "--max_files=10"
  env:
    normal:
      AWS_DEFAULT_REGION: us-east-1
```

Run install 
```bash
helm install test deploy/helm/s3-explorer/ -f your-values.yaml
```

How to install with credentials as env vars
```bash
helm install test deploy/helm/s3-explorer/ -f your-values.yaml --set configuration.env.secret.AWS_ACCESS_KEY_ID=your-key-id --set configuration.env.secret.AWS_SECRET_ACCESS_KEY=your-key
```

Even though AWS_SECRET_ACCESS_KEY and AWS_ACCESS_KEY_ID will be stored as secret with type=Opaque its the wrong way to deal with credentials. 

Please consider using IAM instance profiles or some secrets storage. 