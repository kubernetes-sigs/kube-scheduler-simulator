<template>
  <v-dialog v-model="data.dialog" width="500">
    <template #activator="{ on }">
      <v-btn color="ma-2" v-on="on"> Import </v-btn>
    </template>

    <v-card>
      <v-card-title class="text-h5 grey lighten-2"> Import </v-card-title>

      <v-card-text>
        Import resources and scheduler configuration.<br />
        Note that all current created resources will be deleted and then
        resources are imported..
        <form>
          <input type="file" accept=".yml" @change="readfile" />
        </form>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="green darken-1" text @click="data.dialog = false">
          Cancel
        </v-btn>
        <v-btn
          color="green darken-1"
          text
          :disabled="data.isImportButtonDisabled"
          @click="ImportScheduler()"
        >
          Import
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, inject, reactive } from "@nuxtjs/composition-api";
import { importScheduler, ResourcesForImport } from "~/api/v1/export";
import yaml from "js-yaml";
import SnackBarStoreKey from "../StoreKey/SnackBarStoreKey";
import PriorityClassStoreKey from "../StoreKey/PriorityClassStoreKey";
import StorageClassStoreKey from "../StoreKey/StorageClassStoreKey";
import PersistentVolumeClaimStoreKey from "../StoreKey/PVCStoreKey";
import PersistentVolumeStoreKey from "../StoreKey/PVStoreKey";
import NodeStoreKey from "../StoreKey/NodeStoreKey";
import PodStoreKey from "../StoreKey/PodStoreKey";

interface SelectedItem {
  dialog: boolean;
  filedata: ResourcesForImport;
  isImportButtonDisabled: boolean;
}

export default defineComponent({
  setup() {
    const data = reactive({
      dialog: false,
      isImportButtonDisabled: true,
    } as SelectedItem);
    const priorityclassstore = inject(PriorityClassStoreKey);
    if (!priorityclassstore) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }
    const storageclassstore = inject(StorageClassStoreKey);
    if (!storageclassstore) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }
    const pvcstore = inject(PersistentVolumeClaimStoreKey);
    if (!pvcstore) {
      throw new Error(`${PersistentVolumeClaimStoreKey} is not provided`);
    }
    const pvstore = inject(PersistentVolumeStoreKey);
    if (!pvstore) {
      throw new Error(`${PersistentVolumeStoreKey} is not provided`);
    }
    const nstore = inject(NodeStoreKey);
    if (!nstore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }
    const pstore = inject(PodStoreKey);
    if (!pstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }
    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };
    const ImportScheduler = async () => {
      importScheduler(data.filedata as ResourcesForImport)
        .then(() => {
          priorityclassstore.fetchlist();
          storageclassstore.fetchlist();
          pvcstore.fetchlist();
          pvstore.fetchlist();
          nstore.fetchlist();
          pstore.fetchlist();
        })
        .catch((e) => setServerErrorMessage(e))
        .finally(() => {
          data.dialog = false;
        });
    };
    function readfile(e: { target: { files: FileList | null } }) {
      if (e.target.files === null) return;
      const file = e.target.files[0];
      const reader = new FileReader();

      reader.onload = function () {
        try {
          const filedate: ResourcesForImport = yaml.load(
            reader.result as string
          );
          data.filedata = filedate;
          data.isImportButtonDisabled = false;
        } catch (e) {
          setServerErrorMessage("Failed to load the selected file.");
        }
      };
      reader.readAsText(file);
    }

    return {
      data,
      ImportScheduler,
      readfile,
    };
  },
});
</script>
