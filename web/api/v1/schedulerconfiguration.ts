import { instance } from "@/api/v1/index";
import { SchedulerConfiguration } from "./types";

export const applySchedulerConfiguration = async (
  req: SchedulerConfiguration,
  onError: (_msg: string) => void
) => {
  try {
    const res = await instance.post<SchedulerConfiguration>(
      `/schedulerconfiguration`,
      req
    );
    return res.data;
  } catch (e) {
    onError(e);
  }
};

export const getSchedulerConfiguration = async () => {
  const res = await instance.get<SchedulerConfiguration>(
    `/schedulerconfiguration`
  );
  return res.data;
};
