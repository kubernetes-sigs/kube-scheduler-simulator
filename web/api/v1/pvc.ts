import {
  V1PersistentVolumeClaim,
  V1PersistentVolumeClaimList,
} from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPersistentVolumeClaim = async (
  req: V1PersistentVolumeClaim,
  onError: (_: string) => void
) => {
  try {
    const res = await instance.post<V1PersistentVolumeClaim>(
      `/persistentvolumeclaims`,
      req
    );
    return res.data;
  } catch (e) {
    onError(e);
  }
};

export const listPersistentVolumeClaim = async () => {
  const res = await instance.get<V1PersistentVolumeClaimList>(
    `/persistentvolumeclaims`,
    {}
  );
  return res.data;
};

export const getPersistentVolumeClaim = async (name: string) => {
  const res = await instance.get<V1PersistentVolumeClaim>(
    `/persistentvolumeclaims/${name}`,
    {}
  );
  return res.data;
};

export const deletePersistentVolumeClaim = async (name: string) => {
  const res = await instance.delete(`/persistentvolumeclaims/${name}`, {});
  return res.data;
};
