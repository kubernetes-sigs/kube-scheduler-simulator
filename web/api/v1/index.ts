import axios from "axios";

export const baseURL = process.env.BASE_URL + "/api/v1";
export const instance = axios.create({
  baseURL,
  withCredentials: true,
});

const namespace = "default";
export const namespaceURL = "namespaces/" + namespace;

export const k8sBaseURL = process.env.KUBE_API_SERVER_URL + "/api/v1/";
export const k8sInstance = axios.create({
  baseURL: k8sBaseURL,
  withCredentials: true,
});

export const k8sSchedulingBaseURL =
  process.env.KUBE_API_SERVER_URL + "/apis/scheduling.k8s.io/v1/";
export const k8sSchedulingInstance = axios.create({
  baseURL: k8sSchedulingBaseURL,
  withCredentials: true,
});

export const k8sStorageBaseURL =
  process.env.KUBE_API_SERVER_URL + "/apis/storage.k8s.io/v1/";
export const k8sStorageInstance = axios.create({
  baseURL: k8sStorageBaseURL,
  withCredentials: true,
});
