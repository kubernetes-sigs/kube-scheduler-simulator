import { V1Node, V1NodeList } from "@kubernetes/client-node";
import { k8sInstance } from "@/api/v1/index";
import axios from "axios";

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
    if (axios.isAxiosError(e) && e.response && e.response.status === 404) {
      const res = await createNode(req, onError);
      return res;
    }
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

const createNode = async (req: V1Node, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.post<V1Node>(
      `/nodes?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError("failed to createNode: " + e);
  }
};
