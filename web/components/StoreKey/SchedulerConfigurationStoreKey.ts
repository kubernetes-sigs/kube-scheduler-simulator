import { InjectionKey } from "@nuxtjs/composition-api";
import { SchedulerConfigurationStore } from "../../store/schedulerconfiguration";

const SchedulerConfigurationStoreKey: InjectionKey<SchedulerConfigurationStore> =
  Symbol("SchedulerConfigurationStore");
export default SchedulerConfigurationStoreKey;
