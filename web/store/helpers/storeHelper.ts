import { V1ObjectMeta } from "@kubernetes/client-node";

// resourceObject is an interface of stored resource.
interface resourceObject {
  metadata?: V1ObjectMeta;
}

// ResourceState represents a type of each resource's state.
type ResourceState<T extends resourceObject> = Array<T>;

// createResourceState returns a new resource store's state.
export function createResourceState<T extends resourceObject>(
  resource: T[]
): ResourceState<T> {
  let result: ResourceState<T> = [];
  resource.forEach((r) => {
    result = addResourceToState(result, r);
  });
  return result;
}

// addResourceToState adds the resource to the state.
export function addResourceToState<T extends resourceObject>(
  state: ResourceState<T>,
  r: T
): ResourceState<T> {
  state.push(r);
  return state;
}

// addResourceToState updates the specified resource in the state.
export function modifyResourceInState<T extends resourceObject>(
  state: ResourceState<T>,
  r: T
): ResourceState<T> {
  const i = state.findIndex((res) => res.metadata?.uid === r.metadata?.uid);
  // the resource doesn't exist in the state
  if (i === -1) {
    console.warn("resource doesn't exist in the state");
    return addResourceToState(state, r);
  }
  state.splice(i, 1);
  return addResourceToState(state, r);
}

// deleteResourceInState deletes the specified resouce from the state.
export function deleteResourceInState<T extends resourceObject>(
  state: ResourceState<T>,
  r: T
): ResourceState<T> {
  const i = state.findIndex((res) => res.metadata?.uid === r.metadata?.uid);
  if (i === -1) {
    console.warn("resource doesn't exist in the state");
    return state;
  }
  state.splice(i, 1);
  return state;
}
