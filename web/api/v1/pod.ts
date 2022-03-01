import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { k8sInstance, namespaceURL } from "@/api/v1/index";
import axios from "axios";

export const applyPod = async (req: V1Pod, onError: (_: string) => void) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    const res = await k8sInstance.patch<V1Pod>(
      namespaceURL + `/pods/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/strategic-merge-patch+json" } }
    );
    return res.data;
  } catch (e: any) {
    if (axios.isAxiosError(e) && e.response && e.response.status === 404) {
      const res = await createPod(req, onError);
      return res;
    }
    onError("Caused by applyPod: " + e);
  }
};

export const listPod = async (onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.get<V1PodList>(namespaceURL + `/pods`, {});
    return res.data;
  } catch (e: any) {
    onError("Caused by listPod: " + e);
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
    onError("Caused by getPod: " + e);
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
    onError("Caused by deletePod: " + e);
  }
};

const createPod = async (req: V1Pod, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.post<V1Pod>(
      namespaceURL + `/pods?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError("Caused by createPod: " + e);
  }
};
