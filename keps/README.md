## Internal Kubernetes Enhancement Proposals (KEPs)

This directory contains internal KEP for kube-scheduler-simulator.
The operation is the same as the original KEP for Kubernetes repo, except that it is managed internally in this repository.

> A Kubernetes Enhancement Proposal (KEP) is a way to propose, communicate and coordinate on new efforts for the Kubernetes project. You can read the full details of the project in KEP-0000.
This process is still in a beta state and is mandatory for all enhancements beginning release 1.14.
> https://github.com/kubernetes/enhancements/tree/master/keps

## What is KEP?

Please see https://github.com/kubernetes/enhancements/tree/master/keps.

## How to start to write new KEP?

Follow the process outlined in the KEP template.

https://github.com/kubernetes/enhancements/tree/master/keps/NNNN-kep-template

## What kind of features needs KEP? 

The big change requires KEP. 
Yes, what you want to say is "what is the big change then?", right? 

It's actually case-by-case, but many of the following changes will require KEP, for example:
- Introduce a new CRD.
- Add a new API to a CRD.
- Add a big change to the behavior of the core feature.
- Add a new component on Web UI.

If you are not sure if it is necessary, 
ask [OWNERS](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/master/OWNERS).