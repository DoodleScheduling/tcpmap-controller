# k8stcpmap controller helm chart

Installs the [k8stcpmap-controller](https://github.com/DoodleScheduling/k8stcpmap-controller).

## Installing the Chart

To install the chart with the release name `k8stcpmap-controller`:

```console
helm repo add k8stcpmap-controller https://doodlescheduling.github.io/k8stcpmap-controller/
helm upgrade --install k8stcpmap-controller k8stcpmap-controller/k8stcpmap-controller
```

This command deploys the k8stcpmap-controller with the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Prometheus

The chart comes with a ServiceMonitor/PodMonitor for use with the [Prometheus Operator](https://github.com/coreos/prometheus-operator) which are disabled by default.
If you're not using the Prometheus Operator, you can populate the `podAnnotations` as below:

```yaml
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "metrics"
  prometheus.io/path: "/metrics"
```

## Configuration

See Customizing the Chart Before Installing. To see all configurable options with detailed comments, visit the chart's values.yaml, or run the configuration command:

```sh
$ helm show values k8stcpmap-controller/k8stcpmap-controller
```
