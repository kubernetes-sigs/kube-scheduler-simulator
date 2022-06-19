// LastResourceVersions is used to pass each lastResourceVersion to the server.
export type LastResourceVersions = {
  pods: string;
  nodes: string;
  pvs: string;
  pvcs: string;
  storageClasses: string;
  priorityClasses: string;
};
