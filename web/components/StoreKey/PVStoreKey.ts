import { InjectionKey } from "@nuxtjs/composition-api";
import { PersistentVolumeStore } from "../../store/pv";

const PersistentVolumeStoreKey: InjectionKey<PersistentVolumeStore> =
  Symbol("pvStore");
export default PersistentVolumeStoreKey;
