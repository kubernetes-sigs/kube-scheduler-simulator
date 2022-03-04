import { V1StorageClass, V1StorageClassList } from "@kubernetes/client-node";
import { k8sStorageInstance } from "@/api/v1/index";

export const applyStorageClass = async (
  req: V1StorageClass,
  onError: (_: string) => void
) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    req.kind = "StorageClass";
    req.apiVersion = "storage.k8s.io/v1";
    const res = await k8sStorageInstance.patch<V1StorageClass>(
      `/storageclasses/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    onError("failed to applyStorageClass: " + e);
  }
};

export const listStorageClass = async (onError: (_: string) => void) => {
  try {
    const res = await k8sStorageInstance.get<V1StorageClassList>(
      `/storageclasses`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError("failed to listStorageClass: " + e);
  }
};

export const getStorageClass = async (
  name: string,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sStorageInstance.get<V1StorageClass>(
      `/storageclasses/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError("failed to getStorageClass: " + e);
  }
};

export const deleteStorageClass = async (
  name: string,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sStorageInstance.delete(`/storageclasses/${name}`, {});
    return res.data;
  } catch (e: any) {
    onError("failed to deleteStorageClass: " + e);
  }
};
