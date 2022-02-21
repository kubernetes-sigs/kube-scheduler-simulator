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
    const res = await k8sStorageInstance.patch<V1StorageClass>(
      `/storageclasses/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/strategic-merge-patch+json" } }
    );
    return res.data;
  } catch (e: any) {
    try {
      const res = await createStorageclasses(req, onError);
      return res;
    } catch (e: any) {
      onError(e);
    }
  }
};

export const listStorageClass = async () => {
  const res = await k8sStorageInstance.get<V1StorageClassList>(
    `/storageclasses`,
    {}
  );
  return res.data;
};

export const getStorageClass = async (name: string) => {
  const res = await k8sStorageInstance.get<V1StorageClass>(
    `/storageclasses/${name}`,
    {}
  );
  return res.data;
};

export const deleteStorageClass = async (name: string) => {
  const res = await k8sStorageInstance.delete(`/storageclasses/${name}`, {});
  return res.data;
};

const createStorageclasses = async (
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
    onError(e);
  }
};
