import { reactive } from "@nuxtjs/composition-api";
import { applyNode, deleteNode, listNode, getNode } from "~/api/v1/node";
import { V1Node } from "@kubernetes/client-node";
import { NuxtAxiosInstance } from "@nuxtjs/axios";

type stateType = {
  selectedNode: selectedNode | null;
  nodes: V1Node[];
};

type selectedNode = {
  // isNew represents whether this Node is a new one or not.
  isNew: boolean;
  item: V1Node;
  resourceKind: string;
  isDeletable: boolean;
};

export default function nodeStore(k8sInstance: NuxtAxiosInstance) {
  const state: stateType = reactive({
    selectedNode: null,
    nodes: [],
  });

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

    async fetchlist() {
      const nodes = await listNode(k8sInstance);
      state.nodes = nodes.items;
    },

    async fetchSelected() {
      if (state.selectedNode?.item.metadata?.name && !this.selected?.isNew) {
        const n = await getNode(
          k8sInstance,
          state.selectedNode.item.metadata.name
        );
        this.select(n, false);
      }
    },

    async apply(n: V1Node) {
      await applyNode(k8sInstance, n);
      await this.fetchlist();
    },

    async delete(name: string) {
      await deleteNode(k8sInstance, name);
      await this.fetchlist();
    },
  };
}

export type NodeStore = ReturnType<typeof nodeStore>;
