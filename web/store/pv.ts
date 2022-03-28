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

    async fetchlist() {
      const pvs = await listPersistentVolume();
      state.pvs = pvs.items;
    },

    async fetchSelected() {
      if (
        state.selectedPersistentVolume?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const p = await getPersistentVolume(
          state.selectedPersistentVolume.item.metadata.name
        );
        this.select(p, false);
      }
    },

    async apply(n: V1PersistentVolume) {
      await applyPersistentVolume(n);
      await this.fetchlist();
    },

    async delete(name: string) {
      await deletePersistentVolume(name);
      await this.fetchlist();
    },
  };
}

export type PersistentVolumeStore = ReturnType<typeof pvStore>;
