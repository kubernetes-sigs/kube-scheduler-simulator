<template>
  <PodStoreProvider>
    <NodeStoreProvider>
      <PVStoreProvider>
        <PVCStoreProvider>
          <SchedulerConfigurationStoreProvider>
            <StorageClassStoreProvider>
              <SnackbarStoreProvider>
                <ResourceBar />

                <SchedulerConfigurationEditButton />
                <ResourceAddButton />
                <NodeList />
                <UnscheduledPodList />
                <PVList />
                <PVCList />
                <StorageClassList />
                <Snackbar />
              </SnackbarStoreProvider>
            </StorageClassStoreProvider>
          </SchedulerConfigurationStoreProvider>
        </PVCStoreProvider>
      </PVStoreProvider>
    </NodeStoreProvider>
  </PodStoreProvider>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import SchedulerConfigurationEditButton from "~/components/SchedulerConfigurationEditButton.vue";
import NodeStoreProvider from "~/components/StoreProvider/NodeStoreProvider.vue";
import PVStoreProvider from "~/components/StoreProvider/PVStoreProvider.vue";
import PVCStoreProvider from "~/components/StoreProvider/PVCStoreProvider.vue";
import SchedulerConfigurationStoreProvider from "~/components/StoreProvider/SchedulerConfigurationStoreProvider.vue";
import StorageClassStoreProvider from "~/components/StoreProvider/StorageClassStoreProvider.vue";
import NodeList from "~/components/NodeList.vue";
import UnscheduledPodList from "~/components/UnscheduledPodList.vue";
import PVList from "~/components/PVList.vue";
import PVCList from "~/components/PVCList.vue";
import StorageClassList from "~/components/StorageClassList.vue";
import PodStoreProvider from "~/components/StoreProvider/PodStoreProvider.vue";
import SnackbarStoreProvider from "~/components/StoreProvider/SnackbarStoreProvider.vue";
import ResourceAddButton from "~/components/ResourceAddButton.vue";
import ResourceBar from "~/components/ResourceBar/ResourceBar.vue";
import Snackbar from "~/components/Snackbar.vue";

export default defineComponent({
  components: {
    Snackbar,
    SnackbarStoreProvider,
    NodeStoreProvider,
    PVList,
    PVCList,
    StorageClassList,
    NodeList,
    UnscheduledPodList,
    PodStoreProvider,
    ResourceAddButton,
    ResourceBar,
    StorageClassStoreProvider,
    PVStoreProvider,
    PVCStoreProvider,
    SchedulerConfigurationEditButton,
    SchedulerConfigurationStoreProvider,
  },
});
</script>
