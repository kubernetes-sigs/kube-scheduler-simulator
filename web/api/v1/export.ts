import { V1Namespace, V1Pod } from "@kubernetes/client-node";
import { V1Node } from "@kubernetes/client-node";
import { V1PersistentVolume } from "@kubernetes/client-node";
import { V1PersistentVolumeClaim } from "@kubernetes/client-node";
import { V1StorageClass } from "@kubernetes/client-node";
import { V1PriorityClass } from "@kubernetes/client-node";
import { SchedulerConfiguration } from "./types";
import { AxiosInstance } from "axios";

export default function exportAPI(instance: AxiosInstance) {
  return {
    exportScheduler: async () => {
      const res = await instance.get<ResourcesForImport>(`/export`, {});
      return res.data;
    },

    importScheduler: async (data: ResourcesForImport) => {
      try {
        const res = await instance.post<ResourcesForImport>(`/import`, data);
        return res.data;
      } catch (e: any) {
        throw new Error(e);
      }
    },
  };
}

export declare class ResourcesForImport {
  "pods": V1Pod[];
  "nodes": V1Node[];
  "pvs": V1PersistentVolume[];
  "pvcs": V1PersistentVolumeClaim[];
  "storageClasses": V1StorageClass[];
  "priorityClasses": V1PriorityClass[];
  "schedulerConfig": SchedulerConfiguration;
  "namespaces": V1Namespace[];
}

export type ExportAPI = ReturnType<typeof exportAPI>;
