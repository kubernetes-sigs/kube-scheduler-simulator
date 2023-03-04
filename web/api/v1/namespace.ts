import { V1Namespace, V1NamespaceList } from "@kubernetes/client-node";
import { AxiosInstance} from "axios";

export default function namespaceAPI(k8sInstance: AxiosInstance) {
  return {
    // createNamespace accepts only Namespace that has .metadata.GenerateName.
    // If you want to create a Pod that has .metadata.Name, use applyPod instead.
    createNamespace: async (req: V1Namespace) => {
      try {
        if (!req.metadata?.generateName) {
          throw new Error("metadata.generate")
        }
        req.kind = "Namespace";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.post<V1Namespace>(
          "/namespaces?fieldManager=simulator&force=true",
          req,
          { headers: { "Content-Type": "application/yaml" }}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to create namespace: ${e}`);
      }
    },
    applyNamespace: async (req: V1Namespace) => {
      try {
        if (!req.metadata?.name) {
          throw new Error("metadata.name is not provided.");
        }
        req.kind = "Namespace";
        req.apiVersion = "v1";
        if (req.metadata.managedFields) {
          delete req.metadata.managedFields;
        }
        const res = await k8sInstance.patch<V1Namespace>(
          `/namespaces/${req.metadata.name}?fieldManager=simulator&force=true`,
          req,
          { headers: { "Content-Type": "application/apply-patch+yaml" }}
        );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to apply namespace`);
      }
    },
    listNamespace: async () => {
      try {
        const res = await k8sInstance.get<V1NamespaceList>("/namespaces", {});
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to listt namespaces: ${e}`);
      }
    },
    getNamespace: async (name: string) => {
      try {
        const res = await k8sInstance.get<V1Namespace>(`/namespaces/${name}`, {});
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to get namespace: ${e}`);
      }
    },
    deleteNamespace: async (name: string) => {
      try {
        const res = await k8sInstance.delete<V1Namespace>(`/namespaces/${name}`, {});
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to delete namespace: ${e}`);
      }
    },
    // finalizeNamespace finalizes the specified namespace.
    // This expected to be called when after the deleteNamespace method is called and the namespace's Status remains "Terminating".
    finalizeNamespace: async (req: V1Namespace) => {
      try {
        const res = await k8sInstance.put(`/namespaces/${req.metadata?.name}/finalize`,
        req,
        { headers: { "Content-Type": "application/json" }}
      );
        return res.data;
      } catch (e: any) {
        throw new Error(`failed to finalize namespace: ${e}`);
      }
    }
  };
}
export type NamespaceAPI = ReturnType<typeof namespaceAPI>;
