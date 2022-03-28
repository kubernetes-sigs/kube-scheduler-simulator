import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { k8sInstance, namespaceURL } from "@/api/v1/index";

export const applyPod = async (req: V1Pod) => {
  try {
    if (!req.metadata?.name) {
      throw new Error(`metadata.name is not provided`);
    }
    req.kind = "Pod";
    req.apiVersion = "v1";
    if (req.metadata.managedFields) {
      delete req.metadata.managedFields;
    }
    const res = await k8sInstance.patch<V1Pod>(
      namespaceURL +
        `/pods/${req.metadata.name}?fieldManager=simulator&force=true`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to apply pod: ${e}`);
  }
};

export const listPod = async () => {
  try {
    const res = await k8sInstance.get<V1PodList>(namespaceURL + `/pods`, {});
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to list pods: ${e}`);
  }
};

export const getPod = async (name: string) => {
  try {
    const res = await k8sInstance.get<V1Pod>(
      namespaceURL + `/pods/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to get pod: ${e}`);
  }
};

export const deletePod = async (name: string) => {
  try {
    const res = await k8sInstance.delete(
      namespaceURL + `/pods/${name}?gracePeriodSeconds=0`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to delete pod: ${e}`);
  }
};
