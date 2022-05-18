import { reactive, inject } from "@nuxtjs/composition-api";
import { V1StorageClass } from "@kubernetes/client-node";
import { StorageClassAPIKey } from "~/api/APIProviderKeys";

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

  const storageClassAPI = inject(StorageClassAPIKey);
  if (!storageClassAPI) {
    throw new Error(`${storageClassAPI} is not provided`);
  }

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
      const storageclasses = await storageClassAPI.listStorageClass();
      state.storageclasses = storageclasses.items;
    },

    async apply(n: V1StorageClass) {
      if (n.metadata?.name) {
        await storageClassAPI.applyStorageClass(n);
      } else if (!n.metadata?.name && n.metadata?.generateName) {
        await storageClassAPI.createStorageClass(n);
      } else {
        throw new Error(`
        failed to apply storageclass: storageclass has no metadata.name or metadata.generateName
        `);
      }
      await this.fetchlist();
    },

    async fetchSelected() {
      if (
        state.selectedStorageClass?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const s = await storageClassAPI.getStorageClass(
          state.selectedStorageClass.item.metadata.name
        );
        this.select(s, false);
      }
    },

    async delete(name: string) {
      await storageClassAPI.deleteStorageClass(name);
      await this.fetchlist();
    },
  };
}

export type StorageClassStore = ReturnType<typeof storageclassStore>;
