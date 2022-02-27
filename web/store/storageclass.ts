import { reactive } from "@nuxtjs/composition-api";
import {
  applyStorageClass,
  deleteStorageClass,
  getStorageClass,
  listStorageClass,
} from "~/api/v1/storageclass";
import { V1StorageClass } from "@kubernetes/client-node";

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

export default function storageclassStore() {
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

    async fetchlist(onError: (_: string) => void) {
      const storageclasses = await listStorageClass(onError);
      if (!storageclasses) return;
      state.storageclasses = storageclasses.items;
    },

    async apply(
      n: V1StorageClass,

      onError: (_: string) => void
    ) {
      await applyStorageClass(n, onError);
      await this.fetchlist(onError);
    },

    async fetchSelected(onError: (_: string) => void) {
      if (
        state.selectedStorageClass?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const s = await getStorageClass(
          state.selectedStorageClass.item.metadata.name,
          onError
        );
        if (!s) return;
        this.select(s, false);
      }
    },

    async delete(name: string, onError: (_: string) => void) {
      await deleteStorageClass(name, onError);
      await this.fetchlist(onError);
    },
  };
}

export type StorageClassStore = ReturnType<typeof storageclassStore>;
