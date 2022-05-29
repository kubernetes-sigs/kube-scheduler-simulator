import { V1Node, V1NodeList } from "@kubernetes/client-node";
import { AxiosInstance } from "axios";

export default function nodeAPI(k8sInstance: AxiosInstance) {
  return {
    // createNode accepts only Node that has .metadata.GeneratedName.
    // If you want to create a Node that has .metadata.Name, use applyNode instead.
    createNode: async (req: V1Node) => {
      try {
        if (!req.metadata?.generateName) {
          throw new Error("metadata.generateName is not provided");
        }
        req.kind = "Node";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.post<V1Node>(
          "/nodes?fieldManager=simulator&force=true",
          req,
          { headers: { "Content-Type": "application/yaml" } }
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to create node: ${e}`);
      }
    },
    applyNode: async (req: V1Node) => {
      try {
        if (!req.metadata?.name) {
          throw new Error("metadata.name is not provided");
        }
        req.kind = "Node";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.patch<V1Node>(
          `/nodes/${req.metadata.name}?fieldManager=simulator&force=true`,
          req,
          { headers: { "Content-Type": "application/apply-patch+yaml" } }
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to apply node: ${e}`);
      }
    },

    listNode: async () => {
      try {
        const res = await k8sInstance.get<V1NodeList>("/nodes", {});
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to list nodes: ${e}`);
      }
    },

    getNode: async (name: string) => {
      try {
        const res = await k8sInstance.get<V1Node>(`/nodes/${name}`, {});
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to get node: ${e}`);
      }
    },

    deleteNode: async (name: string) => {
      try {
        const res = await k8sInstance.delete(`/nodes/${name}`, {});
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to delete node: ${e}`);
      }
    },
  };
}
export type NodeAPI = ReturnType<typeof nodeAPI>;
