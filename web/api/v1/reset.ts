import { NuxtAxiosInstance } from "@nuxtjs/axios";

export const reset = async (instance: NuxtAxiosInstance) => {
  const res = await instance.put(`/reset`, {});
  return res.data;
};
