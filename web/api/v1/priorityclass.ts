import { V1PriorityClass, V1PriorityClassList } from "@kubernetes/client-node";
import { k8sSchedulingInstance } from "@/api/v1/index";
import axios from "axios";

export const applyPriorityClass = async (
  req: V1PriorityClass,
  onError: (_: string) => void
) => {
  try {
    if (!req.metadata?.name) {
      onError("metadata.name is not provided");
      return;
    }
    const res = await k8sSchedulingInstance.patch<V1PriorityClass>(
      `/priorityclasses/${req.metadata.name}?fieldManager=simulator`,
      req,
      { headers: { "Content-Type": "application/strategic-merge-patch+json" } }
    );
    return res.data;
  } catch (e: any) {
    if (axios.isAxiosError(e) && e.response && e.response.status === 404) {
      const res = await createPriorityClass(req, onError);
      return res;
    }
    onError("Caused by applyPriorityClass: " + e);
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
    onError("Caused by listPriorityClass: " + e);
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
    onError("Caused by getPriorityClass: " + e);
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
    onError("Caused by deletePriorityClass: " + e);
  }
};

const createPriorityClass = async (
  req: V1PriorityClass,
  onError: (_: string) => void
) => {
  try {
    const res = await k8sSchedulingInstance.post<V1PriorityClass>(
      `/priorityclasses?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError("Caused by createPriorityClass: " + e);
  }
};
