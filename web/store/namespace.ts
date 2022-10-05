import { reactive, inject } from "@nuxtjs/composition-api";
import { V1Namespace } from "@kubernetes/client-node";
import { NamespaceAPIKey } from "~/api/APIProviderKeys";
import {
  createResourceState,
  addResourceToState,
  modifyResourceInState,
  deleteResourceInState,
} from "./helpers/storeHelper";
import { WatchEventType } from "@/types/resources";

type selectedNamespace = {
  isNew: boolean;
  item: V1Namespace;
  resourceKind: string;
  isDeletable: boolean;
}

type stateType = {
  selectedNamespace: selectedNamespace | null;
  namespaces: V1Namespace[];
  lastResourceVersion: string,
}

export default function namespaceStore() {
  const state: stateType = reactive({
    selectedNamespace: null,
    namespaces: [],
    lastResourceVersion: "",
  });

  const namespaceAPI = inject(NamespaceAPIKey);
  if (!namespaceAPI) {
    throw new Error(`${NamespaceAPIKey.description} is not provided`);
  }

  return {
    get namespaces() {
      return state.namespaces;
    },
    get count(): number {
      return state.namespaces.length;
    },
    get selected() {
      return state.selectedNamespace;
    },
    get lastResourceVersion() {
      return state.lastResourceVersion;
    },
    async setLastResourceVersion(pv: V1Namespace) {
      state.lastResourceVersion =
        pv.metadata!.resourceVersion || state.lastResourceVersion;
    },
    select(ns: V1Namespace | null, isNew: boolean) {
      if (ns !== null) {
        state.selectedNamespace = {
          isNew: isNew,
          item: ns,
          resourceKind: "namespace",
          isDeletable: true,
        }
      }
    },
    resetSelected() {
      state.selectedNamespace = null;
    },
    async fetchSelected() {
      if (
        state.selectedNamespace?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const ns = await namespaceAPI.getNamespace(
          state.selectedNamespace.item.metadata.name
        );
        this.select(ns, false);
      }
    },
    async apply(ns: V1Namespace) {
      if (ns.metadata?.name) {
        await namespaceAPI.applyNamespace(ns);
      } else if (ns.metadata?.generateName) {
        await namespaceAPI.createNamespace(ns);
      } else {
        throw new Error(
          "failed to apply namespace: namespace should have metadata.name or metadata.generateName"
        );
      }
    },
    async delete(ns: V1Namespace) {
      if (!ns.metadata?.name) {
        throw new Error(
          "failed to delete namespace: node should have metadata.name"
        )
      }
      namespaceAPI.deleteNamespace(ns.metadata.name).then((res: V1Namespace)=>{
        // When deleting a namespace then it still exists, there is the possibility that any finalizers are specified.
        // We expect that this condition would be almost true.
        if (res.status?.phase === "Terminating" ) {
          res.spec!.finalizers = []
          namespaceAPI.finalizeNamespace(res)
        }
      }).catch((e)=> {
        throw new Error(`failed during the delete process`)
      })

    },
    // initList calls list API, and stores current resource data and the lastResourceVersion.
    async initList() {
      const listns = await namespaceAPI.listNamespace();
      state.namespaces = createResourceState<V1Namespace>(listns.items);
      state.lastResourceVersion = listns.metadata?.resourceVersion!;
    },
    // watchEventHandler handles each notified event.
    async watchEventHandler(eventType: WatchEventType, ns: V1Namespace) {
      switch (eventType) {
        case WatchEventType.ADDED: {
          state.namespaces = addResourceToState(state.namespaces, ns);
          break;
        }
        case WatchEventType.MODIFIED: {
          state.namespaces = modifyResourceInState(state.namespaces, ns);
          break;
        }
        case WatchEventType.DELETED: {
          state.namespaces = deleteResourceInState(state.namespaces, ns);
          break;
        }
        default:
          break;
      }
    },
  }
}

export type NamespaceStore = ReturnType<typeof namespaceStore>;
