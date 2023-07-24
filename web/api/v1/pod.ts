import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { AxiosInstance } from "axios";

export default function podAPI(k8sInstance: AxiosInstance) {
  const defaultNamespace = "default"
  return {
    // createPod accepts only Pod that has .metadata.GeneratedName.
    // If you want to create a Pod that has .metadata.Name, use applyPod instead.
    createPod: async (req: V1Pod) => {
      try {
        if (!req.metadata?.generateName) {
          throw new Error("metadata.generateName is not provided");
        }
        if (!req.metadata.namespace) {
          throw new Error("metadata.namespace is not provided");
        }
        req.kind = "Pod";
        req.apiVersion = "v1"; 
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.post<V1Pod>(
          `namespaces/${req.metadata.namespace}/pods?fieldManager=simulator&force=true`,
          req,
          { headers: { "Content-Type": "application/yaml" } }
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to create pod: ${e}`);
      }
    },
    applyPod: async (req: V1Pod) => {
      try {
        if (!req.metadata?.name) {
          throw new Error("metadata.name is not provided");
        }
        if (!req.metadata.namespace) {
          throw new Error("metadata.namespace is not provided");
        }
        req.kind = "Pod";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.patch<V1Pod>(
            `namespaces/${req.metadata.namespace}/pods/${req.metadata.name}?fieldManager=simulator&force=true`,
          req,
          { headers: { "Content-Type": "application/apply-patch+yaml" } }
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to apply pod: ${e}`);
      }
    },
    listPod: async () => {
      try {
        // This URL path could list all pods on each namespace.
        const res = await k8sInstance.get<V1PodList>(
          "pods",
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to list pods: ${e}`);
      }
    },
    getPod: async (namespace: string, name: string) => {
      try {
        const res = await k8sInstance.get<V1Pod>(
          `namespaces/${namespace}/pods/${name}`,
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to get pod: ${e}`);
      }
    },
    deletePod: async (namespace: string, name: string) => {
      try {
        const res = await k8sInstance.delete(
          `namespaces/${namespace}/pods/${name}?gracePeriodSeconds=0`,
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to delete pod: ${e}`);
      }
    },
  };
}

export type PodAPI = ReturnType<typeof podAPI>;
