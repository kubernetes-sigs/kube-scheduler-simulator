import { V1Node, V1NodeList } from "@kubernetes/client-node";
import { k8sInstance } from "@/api/v1/index";

export const applyNode = async (req: V1Node, onError: (_: string) => void) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    const res = await k8sInstance.patch<V1Node>(
      `/nodes/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/strategic-merge-patch+json" } }
    );
    return res.data;
  } catch (e: any) {
    try {
      const res = await createNode(req, onError);
      return res;
    } catch (e: any) {
      onError(e);
    }
  }
};

export const listNode = async (onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.get<V1NodeList>(`/nodes`, {});
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};

export const getNode = async (name: string, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.get<V1Node>(`/nodes/${name}`, {});
    return res.data;
  } catch (e: any) {
    onError(e);
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
    onError(e);
  }
};

const createNode = async (req: V1Node, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.post<V1Node>(
      `/nodes?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};
