import { V1Node, V1NodeList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyNode = async (req: V1Node, onError: (_: string) => void) => {
  try {
    const res = await instance.post<V1Node>(`/nodes`, req);
    return res.data;
  } catch (e) {
    onError(e);
  }
};

export const listNode = async () => {
  const res = await instance.get<V1NodeList>(`/nodes`, {});
  return res.data;
};

export const getNode = async (name: string) => {
  const res = await instance.get<V1Node>(`/nodes/${name}`, {});
  return res.data;
};

export const deleteNode = async (name: string) => {
  const res = await instance.delete(`/nodes/${name}`, {});
  return res.data;
};
