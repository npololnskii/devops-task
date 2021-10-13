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


### Deployment example

