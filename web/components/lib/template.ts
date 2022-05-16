import {
  V1Node,
  V1PersistentVolume,
  V1PersistentVolumeClaim,
  V1Pod,
  V1StorageClass,
  V1PriorityClass,
} from "@kubernetes/client-node";
import yaml from "js-yaml";

export const podTemplate = (): V1Pod => {
  if (process.env.POD_TEMPLATE) {
    const temp = yaml.load(process.env.POD_TEMPLATE);
    temp.metadata.generateName = temp.metadata.generateName;
    return temp;
  }
  return {};
};

export const nodeTemplate = (namesuffix: string): V1Node => {
  if (process.env.NODE_TEMPLATE) {
    const temp = yaml.load(process.env.NODE_TEMPLATE);
    temp.metadata.name = temp.metadata.name + namesuffix;
    return temp;
  }
  return {};
};

export const pvTemplate = (namesuffix: string): V1PersistentVolume => {
  if (process.env.PV_TEMPLATE) {
    const temp = yaml.load(process.env.PV_TEMPLATE);
    temp.metadata.name = temp.metadata.name + namesuffix;
    return temp;
  }
  return {};
};

export const pvcTemplate = (namesuffix: string): V1PersistentVolumeClaim => {
  if (process.env.PVC_TEMPLATE) {
    const temp = yaml.load(process.env.PVC_TEMPLATE);
    temp.metadata.name = temp.metadata.name + namesuffix;
    return temp;
  }
  return {};
};

export const storageclassTemplate = (namesuffix: string): V1StorageClass => {
  if (process.env.SC_TEMPLATE) {
    const temp = yaml.load(process.env.SC_TEMPLATE);
    temp.metadata.name = temp.metadata.name + namesuffix;
    return temp;
  }
  return { provisioner: "" };
};

export const priorityclassTemplate = (namesuffix: string): V1PriorityClass => {
  if (process.env.PC_TEMPLATE) {
    const temp = yaml.load(process.env.PC_TEMPLATE);
    temp.metadata.name = temp.metadata.name + namesuffix;
    return temp;
  }
  return { value: 1000, globalDefault: true };
};
