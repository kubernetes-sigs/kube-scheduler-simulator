import { InjectionKey } from "@nuxtjs/composition-api";
import { PriorityClassStore } from "~/store/priorityclass";

const PriorityClassStoreKey: InjectionKey<PriorityClassStore> =
  Symbol("PriorityClassStore");
export default PriorityClassStoreKey;
