# Simple BFF demo with Cloud Run  

A demo application shows Simple BFF (Backends For Frontends) with [Cloud Run](https://cloud.google.com/run).  
The requests from BFF go through [Serverless VPC Access](https://cloud.google.com/vpc/docs/configure-serverless-vpc-access) and your VPC, internally reach out to Backend APIs.

![architecture](https://storage.googleapis.com/handson-images/simple-bff-image.png)

## How to use
### 1. Preparation

Set your preferred Google Cloud region name.
```shell
export REGION_NAME={{REGION_NAME}}
```

Set your Google Cloud Project ID
```shell
export PROJECT_ID={{PROJECT_ID}}
```

Set your Artifact Registry repository name
```shell
export REPO_NAME={{REPO_NAME}}
```

Set your VPC name
```shell
export VPC_NAME={{VPC_NAME}}
```

Enable Google Cloud APIs
```shell
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  vpcaccess.googleapis.com \
  cloudtrace.googleapis.com
```

### 2. build container images
Note: please make your own [Artifact Registry repo](https://cloud.google.com/artifact-registry/docs/docker/quickstart) in advance, if you don't have it yet.

#### Build Backend image
```shell
git clone git@github.com:kazshinohara/simple-bff-demo.git
```
```shell
cd simple-bff-demo/backend
```
```shell
gcloud builds submit --tag ${REGION_NAME}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/backend:v1
```

#### Build BFF image
```shell
cd ../bff
```
```shell
gcloud builds submit --tag ${REGION_NAME}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/bff:v1
```

### 3. Prepare Serverless VPC Access Connector


Create a subnet in your VPC, which will be used by Serverless VPC Connector.   
You can choose your preferred CIDR range, but it must be /28 and the one which is not used by other resources.

```shell
gcloud compute networks subnets create serverless-subnet-01 \
--network ${VPC_NAME} \
--range 192.168.255.0/28 \
--enable-flow-logs \
--enable-private-ip-google-access \
--region ${REGION_NAME}
```

Create a Serverless VPC Access Connector.
```shell
gcloud compute networks vpc-access connectors create bff-internal \
--region ${REGION_NAME} \
--subnet serverless-subnet-01
```

Confirm if the connector has been created.
```shell
gcloud compute networks vpc-access connectors describe bff-internal \
--region ${REGION_NAME}
```

### 4. Deploy containers to Cloud Run (fully managed)
Set Cloud Run's base configuration.
```bash
gcloud config set run/region ${REGION_NAME}
gcloud config set run/platform managed
```

Deploy Backend A
```bash
gcloud run deploy backend-a \
--image=${REGION_NAME}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/backend:v1 \
--allow-unauthenticated \
--set-env-vars=VERSION=v1,KIND=backend-a \
--ingress internal
```

Deploy Backend B
```bash
gcloud run deploy backend-b \
--image=${REGION_NAME}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/backend:v1 \
--allow-unauthenticated \
--set-env-vars=VERSION=v1,KIND=backend-b \
--ingress internal
```

Deploy Backend C 
```bash
gcloud run deploy backend-c \
--image=${REGION_NAME}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/backend:v1 \
--allow-unauthenticated \
--set-env-vars=VERSION=v1,KIND=backend-c \
--ingress internal
```

Get all of backend's URLs
```bash
export BE_A=$(gcloud run services describe backend-a --format json | jq -r '.status.address.url')
```
```bash
export BE_B=$(gcloud run services describe backend-b --format json | jq -r '.status.address.url')
```
```bash
export BE_C=$(gcloud run services describe backend-c --format json | jq -r '.status.address.url')
```

```bash
gcloud run deploy bff \
--image=${REGION_NAME}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/bff:v1 \
--allow-unauthenticated \
--set-env-vars=VERSION=v1,KIND=bff,BE_A=${BE_A},BE_B=${BE_B},BE_C=${BE_C} \
--vpc-connector bff-internal \
--vpc-egress all-traffic
```

Get BFF's URL
```bash
export BFF_URL=$(gcloud run services describe bff --format json | jq -r '.status.address.url')
```

### 5. Check behavior
If you could see the following output, it indicates that BFF talks with Backends via the connector.
```shell
curl -X GET ${BFF_URL}/bff | jq
```
```shell
{
  "backend_a_version": "v1",
  "backend_b_version": "v1",
  "backend_c_version": "v1"
}
```

In the end, let's see tracing information via [Cloud Console](https://console.cloud.google.com/traces/list).  
This sample application has [Cloud Trace](https://cloud.google.com/trace) integration, you can see the span between bff and backends like below.
![Trace_list](https://storage.googleapis.com/handson-images/simple-bff-trace.png)


