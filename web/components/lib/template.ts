import {
  V1Node,
  V1PersistentVolume,
  V1PersistentVolumeClaim,
  V1Pod,
  V1StorageClass,
} from '@kubernetes/client-node'
import yaml from 'js-yaml'

export const podTemplate = (namesuffix: string): V1Pod => {
  if (process.env.POD_TEMPLATE) {
    const temp = yaml.load(process.env.POD_TEMPLATE)
    temp.metadata.name = temp.metadata.name + namesuffix
    return temp
  }
  return {}
}

export const nodeTemplate = (namesuffix: string): V1Node => {
  if (process.env.NODE_TEMPLATE) {
    const temp = yaml.load(process.env.NODE_TEMPLATE)
    temp.metadata.name = temp.metadata.name + namesuffix
    return temp
  }
  return {}
}

export const pvTemplate = (namesuffix: string): V1PersistentVolume => {
  if (process.env.PV_TEMPLATE) {
    const temp = yaml.load(process.env.PV_TEMPLATE)
    temp.metadata.name = temp.metadata.name + namesuffix
    return temp
  }
  return {}
}

export const pvcTemplate = (namesuffix: string): V1PersistentVolumeClaim => {
  if (process.env.PVC_TEMPLATE) {
    const temp = yaml.load(process.env.PVC_TEMPLATE)
    temp.metadata.name = temp.metadata.name + namesuffix
    return temp
  }
  return {}
}

export const storageclassTemplate = (namesuffix: string): V1StorageClass => {
  if (process.env.SC_TEMPLATE) {
    const temp = yaml.load(process.env.SC_TEMPLATE)
    temp.metadata.name = temp.metadata.name + namesuffix
    return temp
  }
  return { provisioner: '' }
}
