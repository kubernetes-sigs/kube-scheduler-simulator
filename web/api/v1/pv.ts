import {
  V1PersistentVolume,
  V1PersistentVolumeList,
} from "@kubernetes/client-node";
import { k8sInstance } from "@/api/v1/index";

export const applyPersistentVolume = async (
  req: V1PersistentVolume,
  onError: (_: string) => void
) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    const res = await k8sInstance.patch<V1PersistentVolume>(
      `/persistentvolumes/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/strategic-merge-patch+json" } }
    );
    return res.data;
  } catch (e: any) {
    try {
      const res = await createPersistentvolumes(req, onError);
      return res;
    } catch (e: any) {
      onError(e);
    }
  }
};

export const listPersistentVolume = async () => {
  const res = await k8sInstance.get<V1PersistentVolumeList>(
    `/persistentvolumes`,
    {}
  );
  return res.data;
};

export const getPersistentVolume = async (name: string) => {
  const res = await k8sInstance.get<V1PersistentVolume>(
    `/persistentvolumes/${name}`,
    {}
  );
  return res.data;
};

export const deletePersistentVolume = async (name: string) => {
  const res = await k8sInstance.delete(`/persistentvolumes/${name}`, {});
  return res.data;
};

const createPersistentvolumes = async (
  req: V1PersistentVolume,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sInstance.post<V1PersistentVolume>(
      `/persistentvolumes?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};
