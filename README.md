# k8stcpmap-controller

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4787/badge)](https://bestpractices.coreinfrastructure.org/projects/4787)
[![e2e](https://github.com/DoodleScheduling/k8stcpmap-controller/workflows/e2e/badge.svg)](https://github.com/DoodleScheduling/k8stcpmap-controller/actions)
[![report](https://goreportcard.com/badge/github.com/DoodleScheduling/k8stcpmap-controller)](https://goreportcard.com/report/github.com/DoodleScheduling/k8stcpmap-controller)
[![license](https://img.shields.io/github/license/DoodleScheduling/k8stcpmap-controller.svg)](https://github.com/DoodleScheduling/k8stcpmap-controller/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/DoodleScheduling/k8stcpmap-controller/all.svg)](https://github.com/DoodleScheduling/k8stcpmap-controller/releases)

The k8stcpmap-controller can automatically bind kubernetes services to the nginx ingress tcp proxy.
Using a CRD called TCPIngressMapping you can define which service and what port should be proxied through the nginx ingress controller.
The controller automatically elects a free port which will be exposed on nginx.

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
This is useful if you don't want to set a default nginx ingress and have multiple nginx ingresses.

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

![graph](https://github.com/DoodleScheduling/k8stcpmap-controller/blob/master/docs/graph.jpg?raw=true)


## Helm chart

Please see [chart/k8stcpmap-controller](https://github.com/DoodleScheduling/k8stcpmap-controller) for the helm chart docs.

## Configure the controller

You may change base settings for the controller using env variables (or alternatively command line arguments).
Available env variables:

| Name  | Description | Default |
|-------|-------------| --------|
| `METRICS_ADDR` | The address of the metric endpoint binds to. | `:9556` |
| `PROBE_ADDR` | The address of the probe endpoints binds to. | `:9557` |
| `ENABLE_LEADER_ELECTION` | Enable leader election for controller manager. | `false` |
| `LEADER_ELECTION_NAMESPACE` | Change the leader election namespace. This is by default the same where the controller is deployed. | `` |
| `NAMESPACES` | The controller listens by default for all namespaces. This may be limited to a comma delimited list of dedicated namespaces. | `` |
| `CONCURRENT` | The number of concurrent reconcile workers.  | `1` |
| `TCP_CONFIG_MAP` | The default tcp configmap, accepts `namespace/name` as value to refer to a configmap| |
| `FRONTEND_SERVICE` | The default frontend service, accepts `namespace/name` as value to refer to a service | |
