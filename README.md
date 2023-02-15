# k8stcpmap-controller

[![release](https://img.shields.io/github/release/DoodleScheduling/k8stcpmap-controller/all.svg)](https://github.com/DoodleScheduling/k8stcpmap-controller/releases)
[![release](https://github.com/doodlescheduling/k8stcpmap-controller/actions/workflows/release.yaml/badge.svg)](https://github.com/doodlescheduling/k8stcpmap-controller/actions/workflows/release.yaml)
[![report](https://goreportcard.com/badge/github.com/DoodleScheduling/k8stcpmap-controller)](https://goreportcard.com/report/github.com/DoodleScheduling/k8stcpmap-controller)
[![Coverage Status](https://coveralls.io/repos/github/DoodleScheduling/k8stcpmap-controller/badge.svg?branch=master)](https://coveralls.io/github/DoodleScheduling/k8stcpmap-controller?branch=master)
[![license](https://img.shields.io/github/license/DoodleScheduling/k8stcpmap-controller.svg)](https://github.com/DoodleScheduling/k8stcpmap-controller/blob/master/LICENSE)


The k8stcpmap-controller can automatically bind kubernetes services to the nginx ingress controller tcp proxy.
Using a resource named TCPIngressMapping you can define which service and what port should be proxied through the nginx ingress controller.
The controller automatically elects a free port which will be exposed on nginx as well as the nginx front service.

## Requirements

You need an nginx ingress controller which loads a configmap with port mappings `--tcp-services-configmap`.
See https://kubernetes.github.io/ingress-nginx/user-guide/exposing-tcp-udp-services.
You can deploy an empty configmap.

Example installation of an ingress controller supporting tcp port mapping:
```
kubectl -n ingress-nginx create cm tcp-services-configmap
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm upgrade --wait -i ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --set controller.extraArgs.tcp-services-configmap="\$(POD_NAMESPACE)/tcp-services-configmap"
```

## Use cases
If you have tcp services in your cluster which you need to expose via a single entrypoint.
There is no requirement to port-forward through the api server this way.


## Example TCPIngressMapping

```yaml
apiVersion: networking.infra.doodle.com/v1beta1
kind: TCPIngressMapping
metadata:
  name: mongodb
  namespace: default
spec:
  backendService:
    name: mongodb-primary
    port: mongodb
status:
  electedPort: 10001
```

Its also possible to define for which nginx proxy you want to deploy the proxied port.
This is useful if you don't want to set a default nginx ingress (See env variables) and have multiple nginx ingresses.

```yaml
apiVersion: networking.infra.doodle.com/v1beta1
kind: TCPIngressMapping
metadata:
  name: mongodb
  namespace: default
spec:
  backendService:
    name: mongodb-primary
    port: mongodb
  frontendService:
    name: ingress-nginx-controller
    namespace: ingress-nginx
  tcpConfigMap:
    name: tcp-services-configmap
    namespace: ingress-nginx
```

## Installation

### Helm

Please see [chart/k8stcpmap-controller](https://github.com/DoodleScheduling/k8stcpmap-controller/tree/master/chart/k8stcpmap-controller) for the helm chart docs.

### Manifests/kustomize

Alternatively you may get the bundled manifests in each release to deploy it using kustomize or use them directly.


## Configure the controller

You may change base settings for the controller using env variables (or alternatively command line arguments).
Available env variables:

| Name  | Description | Default |
|-------|-------------| --------|
| `METRICS_ADDR` | The address of the metric endpoint binds to. | `:9556` |
| `PROBE_ADDR` | The address of the probe endpoints binds to. | `:9557` |
| `ENABLE_LEADER_ELECTION` | Enable leader election for controller manager. | `false` |
| `LEADER_ELECTION_NAMESPACE` | Change the leader election namespace. This is by default the same where the controller is deployed. | `` |
| `NAMESPACES` | The controller listens by default for all namespaces. This may be limited to a comma delimted list of dedicated namespaces. | `` |
| `CONCURRENT` | The number of concurrent reconcile workers.  | `1` |
| `TCP_CONFIG_MAP` | The default tcp configmap, accepts `namespace/name` as value to refer to a configmap| |
| `FRONTEND_SERVICE` | The default frontend service, accepts `namespace/name` as value to refer to a service | |
