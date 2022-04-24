import { V1Node, V1NodeList } from "@kubernetes/client-node";
import { NuxtAxiosInstance } from "@nuxtjs/axios";

export const applyNode = async (
  k8sInstance: NuxtAxiosInstance,
  req: V1Node
) => {
  try {
    if (!req.metadata?.name) {
      throw new Error(`metadata.name is not provided`);
    }
    req.kind = "Node";
    req.apiVersion = "v1";
    if (req.metadata.managedFields) {
      delete req.metadata.managedFields;
    }
    const res = await k8sInstance.patch<V1Node>(
      `/nodes/${req.metadata.name}?fieldManager=simulator&force=true`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to apply node: ${e}`);
  }
};

export const listNode = async (k8sInstance: NuxtAxiosInstance) => {
  try {
    const res = await k8sInstance.get<V1NodeList>(`/nodes`, {});
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to list nodes: ${e}`);
  }
};

export const getNode = async (k8sInstance: NuxtAxiosInstance, name: string) => {
  try {
    const res = await k8sInstance.get<V1Node>(`/nodes/${name}`, {});
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to get node: ${e}`);
  }
};

export const deleteNode = async (
  k8sInstance: NuxtAxiosInstance,
  name: string
) => {
  try {
    const res = await k8sInstance.delete(`/nodes/${name}`, {});
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to delete node: ${e}`);
  }
};
