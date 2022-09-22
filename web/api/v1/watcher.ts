import { AxiosInstance } from "axios";
import { LastResourceVersions } from "@/types/api/v1";

export default function watcherAPI(instance: AxiosInstance) {
  return {
    // watchResources is a server push API.
    watchResources: async (lrvs: LastResourceVersions) => {
      try {
        const queries = `podsLastResourceVersion=${lrvs.pods}&nodesLastResourceVersion=${lrvs.nodes}&pvsLastResourceVersion=${lrvs.pvs}&pvcsLastResourceVersion=${lrvs.pvcs}&scsLastResourceVersion=${lrvs.storageClasses}&pcsLastResourceVersion=${lrvs.priorityClasses}`;
        // return stream of Node events.
        return await fetch(
          `${instance.defaults.baseURL}/listwatchresources?${queries}`
        );
      } catch (e: any) {
        throw new Error(`failed to start to watch resources: ${e}`);
      }
    },
  };
}

export type WatcherAPI = ReturnType<typeof watcherAPI>;
