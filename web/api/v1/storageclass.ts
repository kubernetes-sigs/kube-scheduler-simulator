import { V1StorageClass, V1StorageClassList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyStorageClass = async (
  req: V1StorageClass,
  onError: (_: string) => void
) => {
  try {
    const res = await instance.post<V1StorageClass>(`/storageclasses`, req);
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};

export const listStorageClass = async () => {
  const res = await instance.get<V1StorageClassList>(`/storageclasses`, {});
  return res.data;
};

export const getStorageClass = async (name: string) => {
  const res = await instance.get<V1StorageClass>(`/storageclasses/${name}`, {});
  return res.data;
};

export const deleteStorageClass = async (name: string) => {
  const res = await instance.delete(`/storageclasses/${name}`, {});
  return res.data;
};
