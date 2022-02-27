import { reactive } from "@nuxtjs/composition-api";
import {
  applyPersistentVolume,
  deletePersistentVolume,
  getPersistentVolume,
  listPersistentVolume,
} from "~/api/v1/pv";
import { V1PersistentVolume } from "@kubernetes/client-node";

type stateType = {
  selectedPersistentVolume: selectedPersistentVolume | null;
  pvs: V1PersistentVolume[];
};

type selectedPersistentVolume = {
  // isNew represents whether this is a new PersistentVolume or not.
  isNew: boolean;
  item: V1PersistentVolume;
  resourceKind: string;
  isDeletable: boolean;
};

export default function pvStore() {
  const state: stateType = reactive({
    selectedPersistentVolume: null,
    pvs: [],
  });

  return {
    get pvs() {
      return state.pvs;
    },

    get count(): number {
      return state.pvs.length;
    },

    get selected() {
      return state.selectedPersistentVolume;
    },

    select(p: V1PersistentVolume | null, isNew: boolean) {
      if (p !== null) {
        state.selectedPersistentVolume = {
          isNew: isNew,
          item: p,
          resourceKind: "PV",
          isDeletable: true,
        };
      }
    },

    resetSelected() {
      state.selectedPersistentVolume = null;
    },

    async fetchlist(onError: (_: string) => void) {
      const pvs = await listPersistentVolume(onError);
      if (!pvs) return;
      state.pvs = pvs.items;
    },

    async fetchSelected(onError: (_: string) => void) {
      if (
        state.selectedPersistentVolume?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const p = await getPersistentVolume(
          state.selectedPersistentVolume.item.metadata.name,
          onError
        );
        if (!p) return;
        this.select(p, false);
      }
    },

    async apply(
      n: V1PersistentVolume,

      onError: (_: string) => void
    ) {
      await applyPersistentVolume(n, onError);
      await this.fetchlist(onError);
    },

    async delete(name: string, onError: (_: string) => void) {
      await deletePersistentVolume(name, onError);
      await this.fetchlist(onError);
    },
  };
}

export type PersistentVolumeStore = ReturnType<typeof pvStore>;
