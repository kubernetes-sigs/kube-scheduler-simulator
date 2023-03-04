import { reactive, inject } from "@nuxtjs/composition-api";
import { V1Node } from "@kubernetes/client-node";
import { NodeAPIKey } from "~/api/APIProviderKeys";
import {
  createResourceState,
  addResourceToState,
  modifyResourceInState,
  deleteResourceInState,
} from "./helpers/storeHelper";
import { WatchEventType } from "@/types/resources";

type stateType = {
  selectedNode: selectedNode | null;
  nodes: V1Node[];
  lastResourceVersion: string;
};

type selectedNode = {
  // isNew represents whether this Node is a new one or not.
  isNew: boolean;
  item: V1Node;
  resourceKind: string;
  isDeletable: boolean;
};

export default function nodeStore() {
  const state: stateType = reactive({
    selectedNode: null,
    nodes: [],
    lastResourceVersion: "",
  });
  const nodeAPI = inject(NodeAPIKey);
  if (!nodeAPI) {
    throw new Error(`${NodeAPIKey.description} is not provided`);
  }
  return {
    get nodes() {
      return state.nodes;
    },

    get count(): number {
      return state.nodes.length;
    },

    get selected() {
      return state.selectedNode;
    },

    select(n: V1Node | null, isNew: boolean) {
      if (n !== null) {
        state.selectedNode = {
          isNew: isNew,
          item: n,
          resourceKind: "Node",
          isDeletable: true,
        };
      }
    },

    resetSelected() {
      state.selectedNode = null;
    },

    async fetchSelected() {
      if (state.selectedNode?.item.metadata?.name && !this.selected?.isNew) {
        const n = await nodeAPI.getNode(state.selectedNode.item.metadata.name);
        this.select(n, false);
      }
    },

    async apply(n: V1Node) {
      if (n.metadata?.name) {
        await nodeAPI.applyNode(n);
      } else if (n.metadata?.generateName) {
        // This Node can be expected to be a newly created Node. So, use `createNode` instead.
        await nodeAPI.createNode(n);
      } else {
        throw new Error(
          "failed to apply node: node should have metadata.name or metadata.generateName"
        );
      }
    },

    async delete(n: V1Node) {
      if (n.metadata?.name) {
        await nodeAPI.deleteNode(n.metadata.name);
      } else {
        throw new Error(
          "failed to delete node: node should have metadata.name"
        )
      }
    },

    // initList calls list API, and stores current resource data and lastResourceVersion.
    async initList() {
      const listnodes = await nodeAPI.listNode();
      state.nodes = createResourceState<V1Node>(listnodes.items);
      state.lastResourceVersion = listnodes.metadata?.resourceVersion!;
    },

    // watchEventHandler handles each notified event.
    async watchEventHandler(eventType: WatchEventType, node: V1Node) {
      switch (eventType) {
        case WatchEventType.ADDED: {
          state.nodes = addResourceToState(state.nodes, node);
          break;
        }
        case WatchEventType.MODIFIED: {
          state.nodes = modifyResourceInState(state.nodes, node);
          break;
        }
        case WatchEventType.DELETED: {
          state.nodes = deleteResourceInState(state.nodes, node);
          break;
        }
        default:
          break;
      }
    },

    get lastResourceVersion() {
      return state.lastResourceVersion;
    },

    async setLastResourceVersion(node: V1Node) {
      state.lastResourceVersion =
        node.metadata!.resourceVersion || state.lastResourceVersion;
    },
  };
}

export type NodeStore = ReturnType<typeof nodeStore>;
