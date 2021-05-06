# k8stcpmap-controller

![graph](https://github.com/DoodleScheduling/k8stcpmap-controller/blob/master/docs/graph.jpg?raw=true)

## Example TCPIngressMapping

```yaml
apiVersion: network.infra.doodle.com/v1beta1
kind: TCPIngressMapping
metadata:
  name: mongodb
  namespace: devbox-raffis-2
spec:
  backendService:
    name: mongodb-primary
    port: mongodb
status:
  electedPort: 10001
```

## Helm chart

Please see [chart/k8stcpmap-controller](https://github.com/DoodleScheduling/k8stcpmap-controller) for the helm chart docs.

## Configure the controller

You may change base settings for the controller using env variables (or alternatively command line arguments).
Available env variables:

| Name  | Description | Default |
|-------|-------------| --------|
| `METRICS_ADDR` | The address of the metric endpoint binds to. | `:9556` |
| `PROBE_ADDR` | The address of the probe endpoints binds to. | `:9557` |
| `ENABLE_LEADER_ELECTION` | Enable leader election for controller manager. | `true` |
| `LEADER_ELECTION_NAMESPACE` | Change the leader election namespace. This is by default the same where the controller is deployed. | `` |
| `NAMESPACES` | The controller listens by default for all namespaces. This may be limited to a comma delimted list of dedicated namespaces. | `` |
| `CONCURRENT` | The number of concurrent reconcile workers.  | `4` |
