import { reset } from "~/api/v1/reset";

export default function resetStore() {
  return {
    async reset() {
      const data = await reset();
      return data;
    },
  };
}

export type ResetStore = ReturnType<typeof resetStore>;
