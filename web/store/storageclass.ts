import { reactive } from "@nuxtjs/composition-api";
import {
  applyStorageClass,
  deleteStorageClass,
  getStorageClass,
  listStorageClass,
} from "~/api/v1/storageclass";
import { V1StorageClass } from "@kubernetes/client-node";
import { NuxtAxiosInstance } from "@nuxtjs/axios";

type stateType = {
  selectedStorageClass: selectedStorageClass | null;
  storageclasses: V1StorageClass[];
};

type selectedStorageClass = {
  // isNew represents whether this is a new StorageClass or not.
  isNew: boolean;
  item: V1StorageClass;
  resourceKind: string;
  isDeletable: boolean;
};

export default function storageclassStore(
  k8sStorageInstance: NuxtAxiosInstance
) {
  const state: stateType = reactive({
    selectedStorageClass: null,
    storageclasses: [],
  });

  return {
    get storageclasses() {
      return state.storageclasses;
    },

    get count(): number {
      return state.storageclasses.length;
    },

    get selected() {
      return state.selectedStorageClass;
    },

    select(n: V1StorageClass | null, isNew: boolean) {
      if (n !== null) {
        state.selectedStorageClass = {
          isNew: isNew,
          item: n,
          resourceKind: "SC",
          isDeletable: true,
        };
      }
    },

    resetSelected() {
      state.selectedStorageClass = null;
    },

    async fetchlist() {
      const storageclasses = await listStorageClass(k8sStorageInstance);
      state.storageclasses = storageclasses.items;
    },

    async apply(n: V1StorageClass) {
      await applyStorageClass(k8sStorageInstance, n);
      await this.fetchlist();
    },

    async fetchSelected() {
      if (
        state.selectedStorageClass?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const s = await getStorageClass(
          k8sStorageInstance,
          state.selectedStorageClass.item.metadata.name
        );
        this.select(s, false);
      }
    },

    async delete(name: string) {
      await deleteStorageClass(k8sStorageInstance, name);
      await this.fetchlist();
    },
  };
}

export type StorageClassStore = ReturnType<typeof storageclassStore>;
