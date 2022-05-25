import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { namespaceURL } from "@/api/v1/index";
import { AxiosInstance } from "axios";

export default function podAPI(k8sInstance: AxiosInstance) {
  return {
    // createPod accepts only Pod that has .metadata.GeneratedName.
    // If you want to create a Pod that has .metadata.Name, use applyPod instead.
    createPod: async (req: V1Pod) => {
      try {
        if (!req.metadata?.generateName) {
          throw new Error("metadata.generateName is not provided");
        }
        req.kind = "Pod";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.post<V1Pod>(
          namespaceURL + "/pods?fieldManager=simulator&force=true",
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
        req.kind = "Pod";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.patch<V1Pod>(
          namespaceURL +
            `/pods/${req.metadata.name}?fieldManager=simulator&force=true`,
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
        const res = await k8sInstance.get<V1PodList>(
          namespaceURL + "/pods",
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to list pods: ${e}`);
      }
    },
    getPod: async (name: string) => {
      try {
        const res = await k8sInstance.get<V1Pod>(
          namespaceURL + `/pods/${name}`,
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to get pod: ${e}`);
      }
    },
    deletePod: async (name: string) => {
      try {
        const res = await k8sInstance.delete(
          namespaceURL + `/pods/${name}?gracePeriodSeconds=0`,
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
