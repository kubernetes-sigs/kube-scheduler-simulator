import { InjectionKey } from "@nuxtjs/composition-api";
import { PersistentVolumeClaimStore } from "../../store/pvc";

const PersistentVolumeClaimStoreKey: InjectionKey<PersistentVolumeClaimStore> =
  Symbol("PersistentVolumeClaimStore");
export default PersistentVolumeClaimStoreKey;
