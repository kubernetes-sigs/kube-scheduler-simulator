<template>
  <PodStoreProvider>
    <NodeStoreProvider>
      <PVStoreProvider>
        <PVCStoreProvider>
          <SchedulerConfigurationStoreProvider>
            <StorageClassStoreProvider>
              <PriorityClassStoreProvider>
                <ResetProvider>
                  <SnackbarStoreProvider>
                    <ResourceBar />
                    <TopBar />
                    <ResourceAddButton />
                    <ResourceViewPanel />
                    <Snackbar />
                  </SnackbarStoreProvider>
                </ResetProvider>
              </PriorityClassStoreProvider>
            </StorageClassStoreProvider>
          </SchedulerConfigurationStoreProvider>
        </PVCStoreProvider>
      </PVStoreProvider>
    </NodeStoreProvider>
  </PodStoreProvider>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import TopBar from "~/components/TopBar/TopBar.vue";
import NodeStoreProvider from "~/components/StoreProvider/NodeStoreProvider.vue";
import PVStoreProvider from "~/components/StoreProvider/PVStoreProvider.vue";
import PVCStoreProvider from "~/components/StoreProvider/PVCStoreProvider.vue";
import SchedulerConfigurationStoreProvider from "~/components/StoreProvider/SchedulerConfigurationStoreProvider.vue";
import StorageClassStoreProvider from "~/components/StoreProvider/StorageClassStoreProvider.vue";
import PriorityClassStoreProvider from "~/components/StoreProvider/PriorityClassStoreProvider.vue";
import ResourceViewPanel from "~/components/ResourceViews/ResourcesViewPanel.vue";
import PodStoreProvider from "~/components/StoreProvider/PodStoreProvider.vue";
import SnackbarStoreProvider from "~/components/StoreProvider/SnackbarStoreProvider.vue";
import ResourceAddButton from "~/components/ResourceAddButton.vue";
import ResourceBar from "~/components/ResourceBar/ResourceBar.vue";
import Snackbar from "~/components/Snackbar.vue";
import ResetProvider from "~/components/StoreProvider/ResetProvider.vue";

export default defineComponent({
  components: {
    Snackbar,
    SnackbarStoreProvider,
    NodeStoreProvider,
    ResourceViewPanel,
    PodStoreProvider,
    ResourceAddButton,
    ResourceBar,
    StorageClassStoreProvider,
    PVStoreProvider,
    PVCStoreProvider,
    TopBar,
    SchedulerConfigurationStoreProvider,
    PriorityClassStoreProvider,
    ResetProvider,
  },
});
</script>
