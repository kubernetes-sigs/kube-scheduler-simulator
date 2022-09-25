import { InjectionKey } from "@nuxtjs/composition-api";
import { PodAPI } from "./v1/pod";
import { NodeAPI } from "./v1/node";
import { ExportAPI } from "./v1/export";
import { PriorityClassAPI } from "./v1/priorityclass";
import { PVAPI } from "./v1/pv";
import { PVCAPI } from "./v1/pvc";
import { ResetAPI } from "./v1/reset";
import { SchedulerconfigurationAPI } from "./v1/schedulerconfiguration";
import { StorageClassAPI } from "./v1/storageclass";
import { WatcherAPI } from "./v1/watcher";
import { NamespaceAPI } from "./v1/namespace";

export const PodAPIKey: InjectionKey<PodAPI> = Symbol("PodAPI");
export const NodeAPIKey: InjectionKey<NodeAPI> = Symbol("NodeAPI");
export const ExportAPIKey: InjectionKey<ExportAPI> = Symbol("ExportAPI");
export const PriorityClassAPIKey: InjectionKey<PriorityClassAPI> =
  Symbol("PriorityClassAPI");
export const PVAPIKey: InjectionKey<PVAPI> = Symbol("PVAPI");
export const PVCAPIKey: InjectionKey<PVCAPI> = Symbol("PVCAPI");
export const ResetAPIKey: InjectionKey<ResetAPI> = Symbol("ResetAPI");
export const SchedulerconfigurationAPIKey: InjectionKey<SchedulerconfigurationAPI> =
  Symbol("SchedulerconfigurationAPI");
export const StorageClassAPIKey: InjectionKey<StorageClassAPI> =
  Symbol("StorageClassAPI");
export const WatcherAPIKey: InjectionKey<WatcherAPI> = Symbol("WatcherAPI");
export const NamespaceAPIKey: InjectionKey<NamespaceAPI> = Symbol("NamespaceAPI");
