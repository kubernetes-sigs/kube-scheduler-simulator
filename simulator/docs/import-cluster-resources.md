### [Beta] Import your real cluster's resources

The simulator can import resources from your cluster.

To use this, you need to follow these two steps
- Set to `true` the `externalImportEnabled` value in the simulator server configuration.
- Set the path of the kubeconfig file of the your cluster to `KubeConfig` value in the Simulator Server Configuration.

```yaml
externalSchedulerEnabled: false
kubeConfig: "/path/to/your-cluster-kubeconfig"
```

See also [simulator/docs/simulator-server-config.md](simulator/docs/simulator-server-config.md).
