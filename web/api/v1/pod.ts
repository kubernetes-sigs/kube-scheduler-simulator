import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { k8sInstance, namespaceURL } from "@/api/v1/index";

export const applyPod = async (req: V1Pod, onError: (_: string) => void) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    req.kind = "Pod";
    req.apiVersion = "v1";
    const res = await k8sInstance.patch<V1Pod>(
      namespaceURL + `/pods/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    onError("failed to applyPod: " + e);
  }
};

export const listPod = async (onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.get<V1PodList>(namespaceURL + `/pods`, {});
    return res.data;
  } catch (e: any) {
    onError("failed to listPod: " + e);
  }
};

export const getPod = async (name: string, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.get<V1Pod>(
      namespaceURL + `/pods/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError("failed to getPod: " + e);
  }
};

export const deletePod = async (name: string, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.delete(
      namespaceURL + `/pods/${name}?gracePeriodSeconds=0`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError("failed to deletePod: " + e);
  }
};
