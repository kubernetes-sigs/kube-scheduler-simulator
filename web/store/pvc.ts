import { reactive } from "@nuxtjs/composition-api";
import {
  applyPersistentVolumeClaim,
  deletePersistentVolumeClaim,
  getPersistentVolumeClaim,
  listPersistentVolumeClaim,
} from "~/api/v1/pvc";
import { V1PersistentVolumeClaim } from "@kubernetes/client-node";

type stateType = {
  selectedPersistentVolumeClaim: selectedPersistentVolumeClaim | null;
  pvcs: V1PersistentVolumeClaim[];
};

type selectedPersistentVolumeClaim = {
  // isNew represents whether this is a new PersistentVolumeClaim or not.
  isNew: boolean;
  item: V1PersistentVolumeClaim;
  resourceKind: string;
};

export default function pvcStore() {
  const state: stateType = reactive({
    selectedPersistentVolumeClaim: null,
    pvcs: [],
  });

  return {
    get pvcs() {
      return state.pvcs;
    },

    get count(): number {
      return state.pvcs.length;
    },

    get selected() {
      return state.selectedPersistentVolumeClaim;
    },

    select(n: V1PersistentVolumeClaim | null, isNew: boolean) {
      if (n !== null) {
        state.selectedPersistentVolumeClaim = {
          isNew: isNew,
          item: n,
          resourceKind: "PVC",
        };
      }
    },

    resetSelected() {
      state.selectedPersistentVolumeClaim = null;
    },

    async fetchlist() {
      state.pvcs = (await listPersistentVolumeClaim()).items;
    },

    async apply(n: V1PersistentVolumeClaim, onError: (_: string) => void) {
      await applyPersistentVolumeClaim(n, onError);
      await this.fetchlist();
    },

    async fetchSelected() {
      if (
        state.selectedPersistentVolumeClaim?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const p = await getPersistentVolumeClaim(
          state.selectedPersistentVolumeClaim.item.metadata.name
        );
        this.select(p, false);
      }
    },

    async delete(name: string) {
      await deletePersistentVolumeClaim(name);
      await this.fetchlist();
    },
  };
}

export type PersistentVolumeClaimStore = ReturnType<typeof pvcStore>;
