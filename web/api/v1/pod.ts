import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { k8sInstance, namespaceURL } from "@/api/v1/index";

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
    try {
      const res = await createPod(req, onError);
      return res;
    } catch (e: any) {
      onError(e);
    }
  }
};

export const listPod = async () => {
  const res = await k8sInstance.get<V1PodList>(namespaceURL + `/pods`, {});
  return res.data;
};

export const getPod = async (name: string) => {
  const res = await k8sInstance.get<V1Pod>(namespaceURL + `/pods/${name}`, {});
  return res.data;
};

export const deletePod = async (name: string) => {
  const res = await k8sInstance.delete(
    namespaceURL + `/pods/${name}?gracePeriodSeconds=0`,
    {}
  );
  return res.data;
};

const createPod = async (req: V1Pod, onError: (_: string) => void) => {
  try {
    const res = await k8sInstance.post<V1Pod>(
      namespaceURL + `/pods?fieldManager=simulator`,
      req
    );
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};
