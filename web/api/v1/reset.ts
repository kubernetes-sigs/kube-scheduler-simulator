import { AxiosInstance } from "axios";

export default function resetAPI(instance: AxiosInstance) {
  return {
    reset: async () => {
      const res = await instance.put(`/reset`, {});
      return res.data;
    },
  };
}

export type ResetAPI = ReturnType<typeof resetAPI>;
