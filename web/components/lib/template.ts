import {
  V1Node,
  V1PersistentVolume,
  V1PersistentVolumeClaim,
  V1Pod,
  V1StorageClass,
  V1PriorityClass,
  V1Namespace,
} from "@kubernetes/client-node";
import yaml from "js-yaml";

export const podTemplate = (): V1Pod => {
  if (process.env.POD_TEMPLATE) {
    const temp = yaml.load(process.env.POD_TEMPLATE);
    return temp;
  }
  return {};
};

export const nodeTemplate = (): V1Node => {
  if (process.env.NODE_TEMPLATE) {
    const temp = yaml.load(process.env.NODE_TEMPLATE);
    return temp;
  }
  return {};
};

export const pvTemplate = (): V1PersistentVolume => {
  if (process.env.PV_TEMPLATE) {
    const temp = yaml.load(process.env.PV_TEMPLATE);
    return temp;
  }
  return {};
};

export const pvcTemplate = (): V1PersistentVolumeClaim => {
  if (process.env.PVC_TEMPLATE) {
    const temp = yaml.load(process.env.PVC_TEMPLATE);
    return temp;
  }
  return {};
};

export const storageclassTemplate = (): V1StorageClass => {
  if (process.env.SC_TEMPLATE) {
    const temp = yaml.load(process.env.SC_TEMPLATE);
    return temp;
  }
  return { provisioner: "" };
};

export const priorityclassTemplate = (): V1PriorityClass => {
  if (process.env.PC_TEMPLATE) {
    const temp = yaml.load(process.env.PC_TEMPLATE);
    return temp;
  }
  return { value: 1000, globalDefault: true };
};

export const namespaceTemplate = (): V1Namespace => {
  if (process.env.NAMESPACE_TEMPLATE) {
    const temp = yaml.load(process.env.NAMESPACE_TEMPLATE);
    return temp;
  }
  return {};
};
