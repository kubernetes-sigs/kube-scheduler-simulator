import { reactive, inject } from "@nuxtjs/composition-api";
import { V1PersistentVolumeClaim } from "@kubernetes/client-node";
import { PVCAPIKey } from "~/api/APIProviderKeys";
import {
  createResourceState,
  addResourceToState,
  modifyResourceInState,
  deleteResourceInState,
} from "./helpers/storeHelper";
import { WatchEventType } from "@/types/resources";

type stateType = {
  selectedPersistentVolumeClaim: selectedPersistentVolumeClaim | null;
  pvcs: V1PersistentVolumeClaim[];
  lastResourceVersion: string;
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
    lastResourceVersion: "",
  });

  const pvcAPI = inject(PVCAPIKey);
  if (!pvcAPI) {
    throw new Error(`${PVCAPIKey.description} is not provided`);
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

    async apply(n: V1PersistentVolumeClaim) {
      if (n.metadata?.name) {
        await pvcAPI.applyPersistentVolumeClaim(n);
      } else if (n.metadata?.generateName) {
        // This PersistentVolumeClaim can be expected to be a newly created PersistentVolumeClaim. So, use `createPersistentVolumeClaim` instead.
        await pvcAPI.createPersistentVolumeClaim(n);
      } else {
        throw new Error(
          "failed to apply persistentvolumeclaim: persistentvolumeclaim should have metadata.name or metadata.generateName"
        );
      }
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
    },

    // initList calls list API, and stores current resource data and lastResourceVersion.
    async initList() {
      const listpvcs = await pvcAPI.listPersistentVolumeClaim();
      state.pvcs = createResourceState<V1PersistentVolumeClaim>(listpvcs.items);
      state.lastResourceVersion = listpvcs.metadata?.resourceVersion!;
    },

    // watchEventHandler handles each notified event.
    async watchEventHandler(
      eventType: WatchEventType,
      pvc: V1PersistentVolumeClaim
    ) {
      switch (eventType) {
        case WatchEventType.ADDED: {
          state.pvcs = addResourceToState(state.pvcs, pvc);
          break;
        }
        case WatchEventType.MODIFIED: {
          state.pvcs = modifyResourceInState(state.pvcs, pvc);
          break;
        }
        case WatchEventType.DELETED: {
          state.pvcs = deleteResourceInState(state.pvcs, pvc);
          break;
        }
        default:
          break;
      }
    },

    get lastResourceVersion() {
      return state.lastResourceVersion;
    },

    async setLastResourceVersion(pvc: V1PersistentVolumeClaim) {
      state.lastResourceVersion =
        pvc.metadata!.resourceVersion || state.lastResourceVersion;
    },
  };
}

export type PersistentVolumeClaimStore = ReturnType<typeof pvcStore>;
