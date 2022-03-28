import { V1PriorityClass, V1PriorityClassList } from "@kubernetes/client-node";
import { k8sSchedulingInstance } from "@/api/v1/index";

export const applyPriorityClass = async (req: V1PriorityClass) => {
  try {
    if (!req.metadata?.name) {
      throw new Error(`metadata.name is not provided`);
    }
    req.kind = "PriorityClass";
    req.apiVersion = "scheduling.k8s.io/v1";
    if (req.metadata.managedFields) {
      delete req.metadata.managedFields;
    }
    const res = await k8sSchedulingInstance.patch<V1PriorityClass>(
      `/priorityclasses/${req.metadata.name}?fieldManager=simulator&force=true`,
      req,
      { headers: { "Content-Type": "application/apply-patch+yaml" } }
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to apply priority class: ${e}`);
  }
};

export const listPriorityClass = async () => {
  try {
    const res = await k8sSchedulingInstance.get<V1PriorityClassList>(
      `/priorityclasses`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to list priority classes: ${e}`);
  }
};

export const getPriorityClass = async (name: string) => {
  try {
    const res = await k8sSchedulingInstance.get<V1PriorityClass>(
      `/priorityclasses/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to get priority class: ${e}`);
  }
};

export const deletePriorityClass = async (name: string) => {
  try {
    const res = await k8sSchedulingInstance.delete(
      `/priorityclasses/${name}`,
      {}
    );
    return res.data;
  } catch (e: any) {
    throw new Error(`failed to delete priority class: ${e}`);
  }
};
