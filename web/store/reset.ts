import { inject } from "@nuxtjs/composition-api";
import { ResetAPIKey } from "~/api/APIProviderKeys";

export default function resetStore() {
  const resetAPI = inject(ResetAPIKey);
  if (!resetAPI) {
    throw new Error(`${resetAPI} is not provided`);
  }
  return {
    async reset() {
      const data = await resetAPI.reset();
      return data;
    },
  };
}

export type ResetStore = ReturnType<typeof resetStore>;
