import { V1StorageClass, V1StorageClassList } from "@kubernetes/client-node";
import { k8sStorageInstance } from "@/api/v1/index";
import axios from "axios";

export const applyStorageClass = async (
  req: V1StorageClass,
  onError: (_: string) => void
) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    const res = await k8sStorageInstance.patch<V1StorageClass>(
      `/storageclasses/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/strategic-merge-patch+json" } }
    );
    return res.data;
  } catch (e: any) {
    if (axios.isAxiosError(e) && e.response && e.response.status === 404) {
      const res = await createStorageClass(req, onError);
      return res;
    }
    onError("Caused by applyStorageClass: " + e);
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
    onError("Caused by listStorageClass: " + e);
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
    onError("Caused by getStorageClass: " + e);
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
    onError("Caused by deleteStorageClass: " + e);
  }
};

const createStorageClass = async (
  req: V1StorageClass,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sStorageInstance.post<V1StorageClass>(
      `/storageclasses?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError("Caused by createStorageClass: " + e);
  }
};
