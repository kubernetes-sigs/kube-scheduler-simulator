import { V1StorageClass, V1StorageClassList } from "@kubernetes/client-node";
import { k8sStorageInstance } from "@/api/v1/index";

export const applyStorageClass = async (req: V1StorageClass) => {
  try {
    if (!req.metadata?.name) {
      throw new Error(`metadata.name is not provided`);
    }
    req.kind = "StorageClass";
    req.apiVersion = "storage.k8s.io/v1";
    if (req.metadata.managedFields) {
      delete req.metadata.managedFields;
    }
    const res = await k8sStorageInstance.patch<V1StorageClass>(
      `/storageclasses/${req.metadata.name}?fieldManager=simulator&force=true`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to apply storage class: ${e}`);
  }
};

export const listStorageClass = async () => {
  try {
    const res = await k8sStorageInstance.get<V1StorageClassList>(
      `/storageclasses`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to list storage classes: ${e}`);
  }
};

export const getStorageClass = async (name: string) => {
  try {
    const res = await k8sStorageInstance.get<V1StorageClass>(
      `/storageclasses/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to get storage class: ${e}`);
  }
};

export const deleteStorageClass = async (name: string) => {
  try {
    const res = await k8sStorageInstance.delete(`/storageclasses/${name}`, {});
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to delete storage class: ${e}`);
  }
};
