## Integrate your scheduler 

There are several ways to integrate your scheduler into the simulator.

Basically you have two options on the table:
- Integrate your scheduler plugins into the scheduler running in the simulator
  - Check out [custom-plugin.md](./custom-plugin.md) 
- Disable the scheduler running in the simulator and use your scheduler instead
  - We call this feature "external scheduler"
  - Check out [external-scheduler.md](./external-scheduler.md)

Also, if you just want to use your `KubeSchedulerConfig` while using default plugins,
you don't need to follow this page. Check out [simulator-server-config.md](./simulator-server-config.md) instead.
