# Kubernetes client-go credential helper for Google Cloud

`client-go` credential helper that generates Kubernetes `ExecCredentials` objects from the GCloud SDK account.

This server is based on the code from https://github.com/sl1pm4t/gcp-exec-creds (thanks Matt Morrison)

It transfroms the former client into a REST API server that can be run as a sidecar of argocd application controller. It avoids having to rebuild the argocd image.
## Overview

Kubernetes provides a pluggable mechanism for getting user credentials for authenticating with the API server.
`k8s.io/client-go` client libraries and tools using it such as `kubectl` and `kubelet` are able to execute an external command to receive user credentials.
This tool reads the client credentials from the pre-authenticated GCloud SDK account on the local system and generates the appropriate `ExecCredential` API object that can be used by the above tools to authenticate.

See (https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins)

## Usage

The server serves a single API `/creds`. It returns the JSON creds for the current user/service account needed to connect to GKE.

Ideally, this is to be run with workload identity. But it can also be run with the envar `GOOGLE_APPLICATION_CREDENTIALS` (see https://cloud.google.com/docs/authentication/production#automatically)

## Build

The API is generated with openAPI v3 definition, using [oapi-codegen](github.com/deepmap/oapi-codegen) to generate the interfaces (`make generate`)

I use [`ko`](https://github.com/google/ko) to build the image. Change the `repository` variable in the Makefile tp ush to you own registry.

## Test

I do not have unit test, only integration test. Testing locally requires that you have downloaded a GCP service account key for the service account you want to test as explained in https://cloud.google.com/docs/authentication/getting-started.

The key is a JSON file. You simply export the var `GOOGLE_APPLICATION_CREDENTIALS` as the path to the JSON file. For instance, if the file is in your current folder :
```
export GOOGLE_APPLICATION_CREDENTIALS=${PWD}/my-sa.json
```

Then run `make test`

## Deployment

You will find an example of deployment in `deployments` folder. I use the workload identity to associate the application controller to an IAM service account.

Watch out, it has to be customized before you can use it.
### Application Controller sidecar

The file `deployments/application-controller.yaml` contains the kustomization patch to add the sidecar to the aplpication controller.

### Workload Identity

The file `deployments/workload-identity.yaml` contains the configConnector to create teh IAM service account as well as configuring the workload identity for the service account running the application controller.
The file `deployments/sa.yaml` contains the patch for the service account to terminate the workload identity binding.

In those files, we consider that argocd is deployed in a GCP project called `argocd-project` and the IAM service account asscoiated with application controller is `app-controller-gke`
## Configuring the target project

### Authorizing application controller

To configure the authentication and authorization to the cluster on a remote GCP project, we simply have to configure the IAMPolicy for the IAM service account bound to application controller.

Let's say that the project hosting the target cluster is `target-project` In the example of `deployments` folder, we would add the role `roles/container.admin` to the service account `app-controller-gke@argocd-project.iam.gserviceaccount.com`

This grants the permission to argocd to manage the GKE clusters hosted in project `target-project`

### Deploying the cluster secret

A request to the API server sidecar will fetch from GCP the token for the service account associated to application controller. As it has been granted cluster admin permission in project IAM, the service account is authorized to manage full cluster.

The container of the application controller does not have curl installed. Therefore I use `python` to send the request to the sidecar.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gke-cluster1
  labels:
    argocd.argoproj.io/secret-type: cluster
type: Opaque
stringData:
  name:  gke-cluster1
  server: https://<gke endpoint>/
  config: |
    {
      "execProviderConfig": {
          "apiVersion": "client.authentication.k8s.io/v1beta1",
          "command": "python3",
          "args": [
            "-c",
            "import urllib.request; print(urllib.request.urlopen('http://localhost:8080/creds').read().decode('utf8'))"
          ]
        },
        "tlsClientConfig": {
          "caData": "XXXX",
          "insecure":false
      }
    }
```
