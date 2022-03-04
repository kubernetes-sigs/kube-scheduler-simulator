import { V1PriorityClass, V1PriorityClassList } from "@kubernetes/client-node";
import { k8sSchedulingInstance } from "@/api/v1/index";

export const applyPriorityClass = async (
  req: V1PriorityClass,
  onError: (_: string) => void
) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    req.kind = "PriorityClass";
    req.apiVersion = "scheduling.k8s.io/v1";
    const res = await k8sSchedulingInstance.patch<V1PriorityClass>(
      `/priorityclasses/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    onError("failed to applyPriorityClass: " + e);
  }
};

export const listPriorityClass = async (onError: (_: string) => void) => {
  try {
    const res = await k8sSchedulingInstance.get<V1PriorityClassList>(
      `/priorityclasses`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError("failed to listPriorityClass: " + e);
  }
};

export const getPriorityClass = async (
  name: string,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sSchedulingInstance.get<V1PriorityClass>(
      `/priorityclasses/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError("failed to getPriorityClass: " + e);
  }
};

export const deletePriorityClass = async (
  name: string,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sSchedulingInstance.delete(
      `/priorityclasses/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    onError("failed to deletePriorityClass: " + e);
  }
};
