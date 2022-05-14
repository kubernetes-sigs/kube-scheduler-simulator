import {
  V1PersistentVolume,
  V1PersistentVolumeList,
} from "@kubernetes/client-node";
import { AxiosInstance } from "axios";

export default function pvAPI(k8sInstance: AxiosInstance) {
  return {
    applyPersistentVolume: async (req: V1PersistentVolume) => {
      try {
        if (!req.metadata?.name) {
          throw new Error(`metadata.name is not provided`);
        }
        req.kind = "PersistentVolume";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.patch<V1PersistentVolume>(
          `/persistentvolumes/${req.metadata.name}?fieldManager=simulator&force=true`,
          req,
          { headers: { "Content-Type": "application/apply-patch+yaml" } }
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to apply persistent volume: ${e}`);
      }
    },

    listPersistentVolume: async () => {
      try {
        const res = await k8sInstance.get<V1PersistentVolumeList>(
          `/persistentvolumes`,
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to list persistent volumes: ${e}`);
      }
    },

    getPersistentVolume: async (name: string) => {
      try {
        const res = await k8sInstance.get<V1PersistentVolume>(
          `/persistentvolumes/${name}`,
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to get persistent volume: ${e}`);
      }
    },

    deletePersistentVolume: async (name: string) => {
      try {
        const res = await k8sInstance.delete(`/persistentvolumes/${name}`, {});
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to delete persistent volume: ${e}`);
      }
    },
  };
}

export type PVAPI = ReturnType<typeof pvAPI>;
