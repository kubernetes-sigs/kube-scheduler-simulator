# Contributing Guidelines

Welcome to Kubernetes. We are excited about the prospect of you joining our [community](https://git.k8s.io/community)! The Kubernetes community abides by the CNCF [code of conduct](code-of-conduct.md). Here is an excerpt:

_As contributors and maintainers of this project, and in the interest of fostering an open and welcoming community, we pledge to respect all people who contribute through reporting issues, posting feature requests, updating documentation, submitting pull requests or patches, and other activities._

- [Contributor License Agreement](https://git.k8s.io/community/CLA.md) Kubernetes projects require that you sign a Contributor License Agreement (CLA) before we can accept your pull requests
- [Kubernetes Contributor Guide](https://git.k8s.io/community/contributors/guide) - Main contributor documentation, or you can just jump directly to the [contributing section](https://git.k8s.io/community/contributors/guide#contributing)
- [Contributor Cheat Sheet](https://git.k8s.io/community/contributors/guide/contributor-cheatsheet) - Common resources for existing developers

## Mentorship

- [Mentoring Initiatives](https://git.k8s.io/community/mentoring) - We have a diverse set of mentorship programs available that are always looking for volunteers!

## Getting Started

For the frontend, please see [README.md](./web/README.md) on ./web dir.

The first step, you have to prepare tools with make.

```bash
make tools
```

Also, you can run lint, format and test with make.

```bash
# test
make test
# lint
make lint
# format
make format
```

see [Makefile](Makefile) for more details.

## files

```
$ ls -1                        
Dockerfile                # Dockerfile for backend API
LICENSE 
Makefile                   
OWNERS
README.md
RELEASE.md
SECURITY.md
SECURITY_CONTACTS
cloudbuild.yaml           # cloudbuild for backend and frontend image.
code-of-conduct.md
config                    # logics for configuration of simulator. Most configuration is passed via environment variables.
docker-compose.yml        # docker-compose to run up backend and frontend.
docs
errors                    # errors are denined here.
go.mod
go.sum
hack                      # scripts
k8sapiserver              # logics to run up kube-apiserver in the simulator.
node                      # all logics for node should be located here.
persistentvolume          # all logics for persistentvolume should be located here.
persistentvolumeclaim     # all logics for persistentvolumeclaim should be located here.
pod                       # all logics for pod should be located here.
priorityclass             # all logics for priorityclass should be located here.
pvcontroller              # logics to run up pv-controller.
scheduler                 # logics for scheduler and plugins.
server                    # logics to serve HTTP server.
simulator.go              # the entry point of backend API.
storageclass              # all logics for storageclass should be located here.
submodules                # submodules 
tools                     # developer tools
util                      
web                       # frontend 
```

## other docs

- [how the simulator works](./docs/how-it-works.md)
- [API Reference](./docs/api.md)
