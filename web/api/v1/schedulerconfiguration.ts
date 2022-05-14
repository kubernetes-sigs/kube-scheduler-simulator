import { AxiosInstance } from "axios";
import { SchedulerConfiguration } from "./types";

export default function schedulerconfigurationAPI(instance: AxiosInstance) {
  return {
    applySchedulerConfiguration: async (req: SchedulerConfiguration) => {
      try {
        const res = await instance.post<SchedulerConfiguration>(
          `/schedulerconfiguration`,
          req
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to apply scheduler configration: ${e}`);
      }
    },

    getSchedulerConfiguration: async () => {
      const res = await instance.get<SchedulerConfiguration>(
        `/schedulerconfiguration`
      );
      return res.data;
    },
  };
}

export type SchedulerconfigurationAPI = ReturnType<
  typeof schedulerconfigurationAPI
>;
