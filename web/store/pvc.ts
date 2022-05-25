import { reactive, inject } from "@nuxtjs/composition-api";
import { V1PersistentVolumeClaim } from "@kubernetes/client-node";
import { PVCAPIKey } from "~/api/APIProviderKeys";

type stateType = {
  selectedPersistentVolumeClaim: selectedPersistentVolumeClaim | null;
  pvcs: V1PersistentVolumeClaim[];
};

type selectedPersistentVolumeClaim = {
  // isNew represents whether this is a new PersistentVolumeClaim or not.
  isNew: boolean;
  item: V1PersistentVolumeClaim;
  resourceKind: string;
  isDeletable: boolean;
};

export default function pvcStore() {
  const state: stateType = reactive({
    selectedPersistentVolumeClaim: null,
    pvcs: [],
  });

  const pvcAPI = inject(PVCAPIKey);
  if (!pvcAPI) {
    throw new Error(`${pvcAPI} is not provided`);
  }

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
          isDeletable: true,
        };
      }
    },

    resetSelected() {
      state.selectedPersistentVolumeClaim = null;
    },

    async fetchlist() {
      const pvcs = await pvcAPI.listPersistentVolumeClaim();
      state.pvcs = pvcs.items;
    },

    async apply(n: V1PersistentVolumeClaim) {
      if (n.metadata?.name) {
        await pvcAPI.applyPersistentVolumeClaim(n);
      } else if (!n.metadata?.name && n.metadata?.generateName) {
        // This PersistentVolumeClaim can be expected to be a newly created PersistentVolumeClaim. So, use `createPersistentVolumeClaim` instead.
        await pvcAPI.createPersistentVolumeClaim(n);
      } else {
        throw new Error(`
        failed to apply persistentvolumeclaim: persistentvolumeclaim should have metadata.name or metadata.generateName
        `);
      }
      await this.fetchlist();
    },

    async fetchSelected() {
      if (
        state.selectedPersistentVolumeClaim?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const p = await pvcAPI.getPersistentVolumeClaim(
          state.selectedPersistentVolumeClaim.item.metadata.name
        );
        this.select(p, false);
      }
    },

    async delete(name: string) {
      await pvcAPI.deletePersistentVolumeClaim(name);
      await this.fetchlist();
    },
  };
}

export type PersistentVolumeClaimStore = ReturnType<typeof pvcStore>;
