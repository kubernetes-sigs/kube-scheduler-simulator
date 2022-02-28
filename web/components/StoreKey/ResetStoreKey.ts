import { InjectionKey } from "@nuxtjs/composition-api";
import { ResetStore } from "~/store/reset";

const ResetStoreKey: InjectionKey<ResetStore> = Symbol("ResetStore");
export default ResetStoreKey;
