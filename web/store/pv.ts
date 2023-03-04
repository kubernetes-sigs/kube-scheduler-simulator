import { reactive, inject } from "@nuxtjs/composition-api";
import { V1PersistentVolume } from "@kubernetes/client-node";
import { PVAPIKey } from "~/api/APIProviderKeys";
import {
  createResourceState,
  addResourceToState,
  modifyResourceInState,
  deleteResourceInState,
} from "./helpers/storeHelper";
import { WatchEventType } from "@/types/resources";

type stateType = {
  selectedPersistentVolume: selectedPersistentVolume | null;
  pvs: V1PersistentVolume[];
  lastResourceVersion: string;
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
    lastResourceVersion: "",
  });

  const pvAPI = inject(PVAPIKey);
  if (!pvAPI) {
    throw new Error(`${PVAPIKey.description} is not provided`);
  }

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

    async fetchSelected() {
      if (
        state.selectedPersistentVolume?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const p = await pvAPI.getPersistentVolume(
          state.selectedPersistentVolume.item.metadata.name
        );
        this.select(p, false);
      }
    },

    async apply(n: V1PersistentVolume) {
      if (n.metadata?.name) {
        await pvAPI.applyPersistentVolume(n);
      } else if (n.metadata?.generateName) {
        // This PersistentVolume can be expected to be a newly created PersistentVolume. So, use `createPersistentVolume` instead.
        await pvAPI.createPersistentVolume(n);
      } else {
        throw new Error(
          "failed to apply persistentvolume: persistentvolume should have metadata.name or metadata.generateName"
        );
      }
    },

    async delete(pv: V1PersistentVolume) {
      if (pv.metadata?.name) {
        await pvAPI.deletePersistentVolume(pv.metadata.name);
      } else {
        throw new Error(
          "failed to delete persistentvolume: persistentvolume should have metadata.name"
        );
      }
    },

    // initList calls list API, and stores current resource data and lastResourceVersion.
    async initList() {
      const listpvs = await pvAPI.listPersistentVolume();
      state.pvs = createResourceState<V1PersistentVolume>(listpvs.items);
      state.lastResourceVersion = listpvs.metadata?.resourceVersion!;
    },

    // watchEventHandler handles each notified event.
    async watchEventHandler(eventType: WatchEventType, pv: V1PersistentVolume) {
      switch (eventType) {
        case WatchEventType.ADDED: {
          state.pvs = addResourceToState(state.pvs, pv);
          break;
        }
        case WatchEventType.MODIFIED: {
          state.pvs = modifyResourceInState(state.pvs, pv);
          break;
        }
        case WatchEventType.DELETED: {
          state.pvs = deleteResourceInState(state.pvs, pv);
          break;
        }
        default:
          break;
      }
    },

    get lastResourceVersion() {
      return state.lastResourceVersion;
    },

    async setLastResourceVersion(pv: V1PersistentVolume) {
      state.lastResourceVersion =
        pv.metadata!.resourceVersion || state.lastResourceVersion;
    },
  };
}

export type PersistentVolumeStore = ReturnType<typeof pvStore>;
