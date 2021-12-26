import { V1PriorityClass, V1PriorityClassList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPriorityClass = async (
  req: V1PriorityClass,
  onError: (_: string) => void
) => {
  try {
    const res = await instance.post<V1PriorityClass>(`/priorityclasses`, req);
    return res.data;
  } catch (e: any) {
    onError(e);
  }
};

export const listPriorityClass = async () => {
  const res = await instance.get<V1PriorityClassList>(`/priorityclasses`, {});
  return res.data;
};

export const getPriorityClass = async (name: string) => {
  const res = await instance.get<V1PriorityClass>(
    `/priorityclasses/${name}`,
    {}
  );
  return res.data;
};

export const deletePriorityClass = async (name: string) => {
  const res = await instance.delete(`/priorityclasses/${name}`, {});
  return res.data;
};
