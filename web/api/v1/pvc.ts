import {
  V1PersistentVolumeClaim,
  V1PersistentVolumeClaimList,
} from "@kubernetes/client-node";
import { AxiosInstance } from "axios";

export default function pvcAPI(k8sInstance: AxiosInstance) {
  return {
    // createPersistentVolumeClaim accepts only PersistentVolumeClaim that has .metadata.GeneratedName.
    // If you want to create a PersistentVolumeClaim that has .metadata.Name, use applyPersistentVolumeClaim instead.
    createPersistentVolumeClaim: async (req: V1PersistentVolumeClaim) => {
      try {
        if (!req.metadata?.generateName) {
          throw new Error("metadata.generateName is not provided");
        }
        if (!req.metadata?.namespace) {
          throw new Error("metadata.namespace is not provided");
        }
        req.kind = "PersistentVolumeClaim";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.post<V1PersistentVolumeClaim>(
            `namespaces/${req.metadata.namespace}/persistentvolumeclaims?fieldManager=simulator&force=true`,
          req,
          { headers: { "Content-Type": "application/yaml" } }
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to create persistent volume claim: ${e}`);
      }
    },
    applyPersistentVolumeClaim: async (req: V1PersistentVolumeClaim) => {
      try {
        if (!req.metadata?.name) {
          throw new Error("metadata.name is not provided");
        }
        if (!req.metadata?.namespace) {
          throw new Error("metadata.namespace is not provided");
        }
        req.kind = "PersistentVolumeClaim";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.patch<V1PersistentVolumeClaim>(
            `namespaces/${req.metadata.namespace}/persistentvolumeclaims/${req.metadata.name}?fieldManager=simulator&force=true`,
          req,
          { headers: { "Content-Type": "application/apply-patch+yaml" } }
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to apply persistent volume claim: ${e}`);
      }
    },

    listPersistentVolumeClaim: async () => {
      try {
        // This URL path could list all pods on each namespace.
        const res = await k8sInstance.get<V1PersistentVolumeClaimList>(
          "persistentvolumeclaims",
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to list persistent volume claims: ${e}`);
      }
    },

    getPersistentVolumeClaim: async (namespace: string, name: string) => {
      try {
        const res = await k8sInstance.get<V1PersistentVolumeClaim>(
          `namespaces/${namespace}/persistentvolumeclaims/${name}`,
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to get persistent volume claim: ${e}`);
      }
    },

    deletePersistentVolumeClaim: async (namespace: string, name: string) => {
      try {
        const res = await k8sInstance.delete(
          `namespaces/${namespace}/persistentvolumeclaims/${name}`,
          {}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to delete persistent volume claim: ${e}`);
      }
    },
  };
}

export type PVCAPI = ReturnType<typeof pvcAPI>;
