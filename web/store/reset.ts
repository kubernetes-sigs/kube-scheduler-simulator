import { reset } from "~/api/v1/reset";
import { NuxtAxiosInstance } from "@nuxtjs/axios";

export default function resetStore(instance: NuxtAxiosInstance) {
  return {
    async reset() {
      const data = await reset(instance);
      return data;
    },
  };
}

export type ResetStore = ReturnType<typeof resetStore>;
