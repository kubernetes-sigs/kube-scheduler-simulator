# Release Guide

1. An issue is proposing a new release with a changelog since the last release.
2. Make sure your repo is clean by git's standards.
3. Tag the repository from the `master` branch (from the `release-1.19` branch for a patch release) and push the tag `VERSION=v0.19.0 git tag -m $VERSION $VERSION; git push origin $VERSION`.
4. An [OWNER](OWNERS) creates a release branch `git checkout -b release-1.19`. (not required for patch releases)
5. Add the prow-job settings for the new release branch [here](https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/kube-scheduler-simulator). 
6. Push the release branch to the kube-scheduler-simulator repo and ensure branch protection is enabled. (not required for patch releases)
7. Publish a draft release using the tag you created in 3.
8. Perform the [image promotion process](https://github.com/kubernetes/k8s.io/tree/main/k8s.gcr.io#image-promoter).
9. Publish release.
10. Email `kubernetes-sig-scheduling@googlegroups.com` to announce the release.

## Notes
See [post-kube-scheduler-simulator-push-images dashboard](https://testgrid.k8s.io/sig-scheduling#post-kube-scheduler-simulator-push-images) for staging registry image build job status.

View the kube-scheduler-simulator staging registry using [this URL](https://console.cloud.google.com/gcr/images/k8s-staging-sched-simulator/GLOBAL) in a web browser
or use the below `gcloud` commands.

List images in staging registry.
```shell
gcloud container images list --repository gcr.io/k8s-staging-sched-simulator
```

List simulator-backend and simulator-frontend image tags in the staging registry.
```shell
gcloud container images list-tags gcr.io/k8s-staging-sched-simulator/simulator-backend
gcloud container images list-tags gcr.io/k8s-staging-sched-simulator/simulator-frontend
```
