import { V1Node, V1NodeList } from "@kubernetes/client-node";
import { k8sInstance } from "@/api/v1/index";

export const applyNode = async (req: V1Node, onError: (_: string) => void) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    req.kind = "Node";
    req.apiVersion = "v1";
    const res = await k8sInstance.patch<V1Node>(
      `/nodes/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    onError("failed to applyNode: " + e);
  }
};

export const listNode = async (onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.get<V1NodeList>(`/nodes`, {});
    return res.data;
  } catch (e: any) {
    onError("failed to listNode: " + e);
  }
};

export const getNode = async (name: string, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.get<V1Node>(`/nodes/${name}`, {});
    return res.data;
  } catch (e: any) {
    onError("failed to getNode: " + e);
  }
};

export const deleteNode = async (
  name: string,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sInstance.delete(`/nodes/${name}`, {});
    return res.data;
  } catch (e: any) {
    onError("failed to deleteNode: " + e);
  }
};
