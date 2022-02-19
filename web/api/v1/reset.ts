import { instance } from "@/api/v1/index";

export const reset = async () => {
  const res = await instance.put(`/reset`, {});
  return res.data;
};
