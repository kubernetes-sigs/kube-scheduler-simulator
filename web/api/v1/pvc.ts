import {
  V1PersistentVolumeClaim,
  V1PersistentVolumeClaimList,
} from "@kubernetes/client-node";
import { k8sInstance, namespaceURL } from "@/api/v1/index";

export const applyPersistentVolumeClaim = async (
  req: V1PersistentVolumeClaim
) => {
  try {
    if (!req.metadata?.name) {
      throw new Error(`metadata.name is not provided`);
    }
    req.kind = "PersistentVolumeClaim";
    req.apiVersion = "v1";
    if (req.metadata.managedFields) {
      delete req.metadata.managedFields;
    }
    const res = await k8sInstance.patch<V1PersistentVolumeClaim>(
      namespaceURL +
        `/persistentvolumeclaims/${req.metadata.name}?fieldManager=simulator&force=true`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to apply persistent volume claim: ${e}`);
  }
};

export const listPersistentVolumeClaim = async () => {
  try {
    const res = await k8sInstance.get<V1PersistentVolumeClaimList>(
      namespaceURL + `/persistentvolumeclaims`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to list persistent volume claims: ${e}`);
  }
};

export const getPersistentVolumeClaim = async (name: string) => {
  try {
    const res = await k8sInstance.get<V1PersistentVolumeClaim>(
      namespaceURL + `/persistentvolumeclaims/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to get persistent volume claim: ${e}`);
  }
};

export const deletePersistentVolumeClaim = async (name: string) => {
  try {
    const res = await k8sInstance.delete(
      namespaceURL + `/persistentvolumeclaims/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to delete persistent volume claim: ${e}`);
  }
};
