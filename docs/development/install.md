## Install

### Requirements
Following tools are required to work on that package.

- [make](https://www.gnu.org/software/make/) - to execute build goals
- [golang](https://golang.org/) - to compile the code
- [kind](https://kind.sigs.k8s.io/) or access to k8s cluster - to deploy and test operator
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - to interact with k8s cluster via CLI
- [kustomize](https://kustomize.io/) - to generate deployment configs
- [kubebuilder](https://book.kubebuilder.io) - framework to build operators
- [operator framework](https://operatorframework.io/) - framework to maintain project structure
- [helm](https://helm.sh/) - to work with helm charts

If you have to build Docker images on your host, 
you also need to have [Docker](https://www.docker.com/) or its alternative installed.

### Prepare environment

If you have access to the docker registry and k8s installation that you can use for development purposes, you may skip
corresponding steps.

Otherwise, create a local instance of k8s.
```
    kind create cluster
    Creating cluster "kind" ...
    ✓ Ensuring node image (kindest/node:v1.20.2) 🖼
    ✓ Preparing nodes 📦
    ✓ Writing configuration 📜
    ✓ Starting control-plane 🕹️
    ✓ Installing CNI 🔌
    ✓ Installing StorageClass 💾
    Set kubectl context to "kind-kind"
    You can now use your cluster with:

    kubectl cluster-info --context kind-kind

    Thanks for using kind! 😊
```

## Install
You can use helm for deploy injector in the cluster.
```
helm install injector ./deploy/helm/injector
```
