<template>
  <v-dialog v-model="data.dialog" width="500">
    <template #activator="{ on }">
      <v-btn class="ma-2" color="error" v-on="on"> Reset </v-btn>
    </template>

    <v-card>
      <v-card-title class="2">
        Are you sure to reset all resources and scheduler configuration?
      </v-card-title>
      <v-divider></v-divider>
      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="green darken-1" text @click="resetFn"> Reset </v-btn>
        <v-btn color="green darken-1" text @click="data.dialog = false">
          Cancel
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { inject, defineComponent, reactive } from "@nuxtjs/composition-api";
import { ResetAPIKey } from "~/api/APIProviderKeys";
import PodStoreKey from "../StoreKey/PodStoreKey";
import NodeStoreKey from "../StoreKey/NodeStoreKey";
import PersistentVolumeStoreKey from "../StoreKey/PVStoreKey";
import PersistentVolumeClaimStoreKey from "../StoreKey/PVCStoreKey";
import StorageClassStoreKey from "../StoreKey/StorageClassStoreKey";
import PriorityClassStoreKey from "../StoreKey/PriorityClassStoreKey";
import SchedulerConfigurationStoreKey from "../StoreKey/SchedulerConfigurationStoreKey";
import SnackBarStoreKey from "../StoreKey/SnackBarStoreKey";

export default defineComponent({
  setup() {
    const resetAPI = inject(ResetAPIKey);
    if (!resetAPI) {
      throw new Error(`${resetAPI} is not provided`);
    }
    const podstore = inject(PodStoreKey);
    if (!podstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const nodestore = inject(NodeStoreKey);
    if (!nodestore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const pvstore = inject(PersistentVolumeStoreKey);
    if (!pvstore) {
      throw new Error(`${pvstore} is not provided`);
    }

    const pvcstore = inject(PersistentVolumeClaimStoreKey);
    if (!pvcstore) {
      throw new Error(`${pvcstore} is not provided`);
    }

    const storageclassstore = inject(StorageClassStoreKey);
    if (!storageclassstore) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }

    const priorityclassstore = inject(PriorityClassStoreKey);
    if (!priorityclassstore) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }

    const schedulerconfigurationstore = inject(SchedulerConfigurationStoreKey);
    if (!schedulerconfigurationstore) {
      throw new Error(`${SchedulerConfigurationStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const setInfoMessage = (message: string) => {
      snackbarstore.setServerInfoMessage(message);
    };

    const resetFn = async () => {
      try {
        await resetAPI.reset();
        await Promise.all([
          nodestore.fetchlist(),
          podstore.fetchlist(),
          pvstore.fetchlist(),
          pvcstore.fetchlist(),
          storageclassstore.fetchlist(),
          priorityclassstore.fetchlist(),
        ]);
        setInfoMessage("Successfully reset all resources");
      } catch (e: any) {
        setServerErrorMessage(e.message);
      } finally {
        data.dialog = false;
      }
    };

    const data = reactive({
      dialog: false,
    });

    return {
      resetFn,
      data,
    };
  },
});
</script>
