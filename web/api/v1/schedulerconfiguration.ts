import { SchedulerConfiguration } from "./types";
import { NuxtAxiosInstance } from "@nuxtjs/axios";

export const applySchedulerConfiguration = async (
  instance: NuxtAxiosInstance,
  req: SchedulerConfiguration
) => {
  try {
    const res = await instance.post<SchedulerConfiguration>(
      `/schedulerconfiguration`,
      req
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to apply scheduler configration: ${e}`);
  }
};

export const getSchedulerConfiguration = async (
  instance: NuxtAxiosInstance
) => {
  const res = await instance.get<SchedulerConfiguration>(
    `/schedulerconfiguration`
  );
  return res.data;
};
