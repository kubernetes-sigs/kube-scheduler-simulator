<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, useContext } from "@nuxtjs/composition-api";
import podAPI from "./v1/pod";
import nodeAPI from "./v1/node";
import priorityClassAPI from "./v1/priorityclass";
import exportAPI from "./v1/export";
import pvAPI from "./v1/pv";
import pvcAPI from "./v1/pvc";
import resetAPI from "./v1/reset";
import schedulerconfigurationAPI from "./v1/schedulerconfiguration";
import storageClassAPI from "./v1/storageclass";
import watcherAPI from "./v1/watcher";
import { PodAPIKey } from "./APIProviderKeys";
import { NodeAPIKey } from "./APIProviderKeys";
import { PriorityClassAPIKey } from "./APIProviderKeys";
import { ExportAPIKey } from "./APIProviderKeys";
import { PVAPIKey } from "./APIProviderKeys";
import { PVCAPIKey } from "./APIProviderKeys";
import { ResetAPIKey } from "./APIProviderKeys";
import { SchedulerconfigurationAPIKey } from "./APIProviderKeys";
import { StorageClassAPIKey } from "./APIProviderKeys";
import { WatcherAPIKey } from "./APIProviderKeys";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(PodAPIKey, podAPI(app.$k8sInstance));
    provide(NodeAPIKey, nodeAPI(app.$k8sInstance));
    provide(PriorityClassAPIKey, priorityClassAPI(app.$k8sSchedulingInstance));
    provide(ExportAPIKey, exportAPI(app.$instance));
    provide(PVAPIKey, pvAPI(app.$k8sInstance));
    provide(PVCAPIKey, pvcAPI(app.$k8sInstance));
    provide(ResetAPIKey, resetAPI(app.$instance));
    provide(
      SchedulerconfigurationAPIKey,
      schedulerconfigurationAPI(app.$instance)
    );
    provide(StorageClassAPIKey, storageClassAPI(app.$k8sStorageInstance));
    provide(WatcherAPIKey, watcherAPI(app.$instance));
    return {};
  },
});
</script>
