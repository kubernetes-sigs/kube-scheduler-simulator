import { Plugin } from "@nuxt/types";

const ApiRuntimeConfigPlugin: Plugin = (context, inject): void => {
  const baseURL = context.$config.baseURL + "/api/v1";
  const instance = context.$axios.create({
    baseURL: baseURL,
    withCredentials: true,
  });

  const k8sBaseURL = context.$config.kubeApiServerURL + "/api/v1/";
  const k8sInstance = context.$axios.create({
    baseURL: k8sBaseURL,
    withCredentials: true,
  });
  const k8sSchedulingBaseURL =
    context.$config.kubeApiServerURL + "/apis/scheduling.k8s.io/v1/";
  const k8sSchedulingInstance = context.$axios.create({
    baseURL: k8sSchedulingBaseURL,
    withCredentials: true,
  });

  const k8sStorageBaseURL =
    context.$config.kubeApiServerURL + "/apis/storage.k8s.io/v1/";
  const k8sStorageInstance = context.$axios.create({
    baseURL: k8sStorageBaseURL,
    withCredentials: true,
  });

  inject("instance", instance);
  inject("k8sInstance", k8sInstance);
  inject("k8sSchedulingInstance", k8sSchedulingInstance);
  inject("k8sStorageInstance", k8sStorageInstance);
};

export default ApiRuntimeConfigPlugin;
