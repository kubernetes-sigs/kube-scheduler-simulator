import { InjectionKey } from "@nuxtjs/composition-api";
import { NamespaceStore } from "../../store/namespace";

const NamespaceStoreKey: InjectionKey<NamespaceStore> = Symbol("NamespaceStore");
export default NamespaceStoreKey;