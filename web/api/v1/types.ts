export interface SchedulerConfiguration {
  kind: string
  apiVersion: string
  Profiles: KubeSchedulerProfile[]
}

export interface KubeSchedulerProfile {
  SchedulerName: string
  Plugins: Plugins
}

export interface Plugins {
  QueueSort: PluginSet
  PreFilter: PluginSet
  Filter: PluginSet
  PostFilter: PluginSet
  PreScore: PluginSet
  Score: PluginSet
  Reserve: PluginSet
  Permit: PluginSet
  PreBind: PluginSet
  Bind: PluginSet
  PostBind: PluginSet
}

export interface PluginSet {
  Enabled: Plugin[]
  Disabled: Plugin[]
}

export interface Plugin {
  Name: string
  Weight: number
}
