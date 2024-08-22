### [Beta] Import your real cluster's resources

There are two ways to import resources from your cluster. These methods cannot be used simultaneously.
- Import resources from your cluster once when initializing the simulator.
- Keep importing resources from your cluster.

#### Import resources once when initializing the simulator

To use this, you need to follow these two steps in the simulator configuration:
- Set `true` to `externalImportEnabled`.
- Set the path of the kubeconfig file for the your cluster to `KubeConfig`. 
  - This feature only requires the read permission for resources.

```yaml
externalImportEnabled: true
kubeConfig: "/path/to/your-cluster-kubeconfig"
```

#### Keep importing resources

To use this, you need to follow these two steps in the scheduler configuration:
- Set `true` to `resourceSyncEnabled`.
- Set the path of the kubeconfig file for the your cluster to `KubeConfig`. 
  - This feature only requires the read permission for resources.

```yaml
resourceSyncEnabled: true
kubeConfig: "/path/to/your-cluster-kubeconfig"
```

> [!NOTE]
> When you enable `resourceSyncEnabled`, adding/updating/deleting resources directly in the simulator cluster could cause a problem of syncing. 
> You can do them for debugging etc purposes though, make sure you reboot the simulator and the fake source cluster afterward.

See [simulator/docs/simulator-server-config.md](simulator/docs/simulator-server-config.md) for more information about the simulator configuration.
