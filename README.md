# tcpmap-controller

[![release](https://img.shields.io/github/release/DoodleScheduling/tcpmap-controller/all.svg)](https://github.com/DoodleScheduling/tcpmap-controller/releases)
[![release](https://github.com/doodlescheduling/tcpmap-controller/actions/workflows/release.yaml/badge.svg)](https://github.com/doodlescheduling/tcpmap-controller/actions/workflows/release.yaml)
[![report](https://goreportcard.com/badge/github.com/DoodleScheduling/tcpmap-controller)](https://goreportcard.com/report/github.com/DoodleScheduling/tcpmap-controller)
[![Coverage Status](https://coveralls.io/repos/github/DoodleScheduling/tcpmap-controller/badge.svg?branch=master)](https://coveralls.io/github/DoodleScheduling/tcpmap-controller?branch=master)
[![license](https://img.shields.io/github/license/DoodleScheduling/tcpmap-controller.svg)](https://github.com/DoodleScheduling/tcpmap-controller/blob/master/LICENSE)


The tcpmap-controller can automatically bind kubernetes services to the nginx ingress controller tcp proxy.
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

Please see [chart/tcpmap-controller](https://github.com/DoodleScheduling/tcpmap-controller/tree/master/chart/tcpmap-controller) for the helm chart docs.

### Manifests/kustomize

Alternatively you may get the bundled manifests in each release to deploy it using kustomize or use them directly.


## Configure the controller

The controller is configurable by cmd args:
```
--concurrent int                            The number of concurrent Pod reconciles. (default 4)
--enable-leader-election                    Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
--frontend-service string                   Set the default nginx controller service. Might be set in the resource itself
--graceful-shutdown-timeout duration        The duration given to the reconciler to finish before forcibly stopping. (default 10m0s)
--health-addr string                        The address the health endpoint binds to. (default ":9557")
--insecure-kubeconfig-exec                  Allow use of the user.exec section in kubeconfigs provided for remote apply.
--insecure-kubeconfig-tls                   Allow that kubeconfigs provided for remote apply can disable TLS verification.
--kube-api-burst int                        The maximum burst queries-per-second of requests sent to the Kubernetes API. (default 300)
--kube-api-qps float32                      The maximum queries-per-second of requests sent to the Kubernetes API. (default 50)
--leader-election-lease-duration duration   Interval at which non-leader candidates will wait to force acquire leadership (duration string). (default 35s)
--leader-election-release-on-cancel         Defines if the leader should step down voluntarily on controller manager shutdown. (default true)
--leader-election-renew-deadline duration   Duration that the leading controller manager will retry refreshing leadership before giving up (duration string). (default 30s)
--leader-election-retry-period duration     Duration the LeaderElector clients should wait between tries of actions (duration string). (default 5s)
--log-encoding string                       Log encoding format. Can be 'json' or 'console'. (default "json")
--log-level string                          Log verbosity level. Can be one of 'trace', 'debug', 'info', 'error'. (default "info")
--max-port int32                            Do not elect a port above (default 65535)
--max-retry-delay duration                  The maximum amount of time for which an object being reconciled will have to wait before a retry. (default 15m0s)
--metrics-addr string                       The address the metric endpoint binds to. (default ":9556")
--min-port int32                            Do not elect a port bellow. (default 1025)
--min-retry-delay duration                  The minimum amount of time for which an object being reconciled will have to wait before a retry. (default 750ms)
--tcp-services-configmap string             Set the default tcp configmap (https://kubernetes.github.io/ingress-nginx/user-guide/exposing-tcp-udp-services/). Might be set in the resource itself.
--watch-all-namespaces                      Watch for resources in all namespaces, if set to false it will only watch the runtime namespace. (default true)
--watch-label-selector string               Watch for resources with matching labels e.g. 'sharding.fluxcd.io/shard=shard1'.
```
