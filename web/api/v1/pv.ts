import {
  V1PersistentVolume,
  V1PersistentVolumeList,
} from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPersistentVolume = async (
  req: V1PersistentVolume,
  onError: (_: string) => void
) => {
  try {
    const res = await instance.post<V1PersistentVolume>(
      `/persistentvolumes`,
      req
    );
    return res.data;
  } catch (e) {
    onError(e);
  }
};

export const listPersistentVolume = async () => {
  const res = await instance.get<V1PersistentVolumeList>(
    `/persistentvolumes`,
    {}
  );
  return res.data;
};

export const getPersistentVolume = async (name: string) => {
  const res = await instance.get<V1PersistentVolume>(
    `/persistentvolumes/${name}`,
    {}
  );
  return res.data;
};

export const deletePersistentVolume = async (name: string) => {
  const res = await instance.delete(`/persistentvolumes/${name}`, {});
  return res.data;
};
