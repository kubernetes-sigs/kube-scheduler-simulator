### [Beta] Import your real cluster's resources

There are two ways to import resources from your cluster.
- Import resources from your cluster once when initializing the simulator.
- Keep importing resources from your cluster.

#### Import resources once when initializing the simulator

To use this, you need to follow these two steps in the simulator configuration:
- Set `true` to `externalImportEnabled`.
- Set the path of the kubeconfig file for the your cluster to `KubeConfig`. 
  - This feature only requires the read permission for resources.

```yaml
externalSchedulerEnabled: true
kubeConfig: "/path/to/your-cluster-kubeconfig"
```

#### Keep importing resources

To use this, you need to follow these two steps in the scheduler configuration:
- Set `true` to `externalImportSynced`. 
- Set the path of the kubeconfig file for the your cluster to `KubeConfig`. 
  - This feature only requires the read permission for resources.

```yaml
externalSchedulerSynced: true
kubeConfig: "/path/to/your-cluster-kubeconfig"
```

See [simulator/docs/simulator-server-config.md](simulator/docs/simulator-server-config.md) for more information about the simulator configuration.