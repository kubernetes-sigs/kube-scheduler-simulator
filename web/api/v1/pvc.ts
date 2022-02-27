import {
  V1PersistentVolumeClaim,
  V1PersistentVolumeClaimList,
} from "@kubernetes/client-node";
import { k8sInstance, namespaceURL } from "@/api/v1/index";

export const applyPersistentVolumeClaim = async (
  req: V1PersistentVolumeClaim,
  onError: (_: string) => void
) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    const res = await k8sInstance.patch<V1PersistentVolumeClaim>(
      namespaceURL +
        `/persistentvolumeclaims/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/strategic-merge-patch+json" } }
    );
    return res.data;
  } catch (e: any) {
    try {
      const res = await createPersistentVolumeClaim(req, onError);
      return res;
    } catch (e: any) {
      onError(e);
    }
  }
};

export const listPersistentVolumeClaim = async (
  onError: (_: string) => void
) => {
  try {
    const res = await k8sInstance.get<V1PersistentVolumeClaimList>(
      namespaceURL + `/persistentvolumeclaims`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};

export const getPersistentVolumeClaim = async (
  name: string,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sInstance.get<V1PersistentVolumeClaim>(
      namespaceURL + `/persistentvolumeclaims/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};

export const deletePersistentVolumeClaim = async (
  name: string,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sInstance.delete(
      namespaceURL + `/persistentvolumeclaims/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};

const createPersistentVolumeClaim = async (
  req: V1PersistentVolumeClaim,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sInstance.post<V1PersistentVolumeClaim>(
      namespaceURL + `/persistentvolumeclaims?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};
