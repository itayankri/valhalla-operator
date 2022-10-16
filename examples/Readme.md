## Environment Setup
### Setting up a kind cluster
Kind is a tool for running a local version of kubernetea for development use. In case you already have an operational kubernetes cluster you can skip this step. If you don't make sure you have [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) installed on your machine.
Another requirement is changing the the "storage-driver" configuration in the docker config. Docker uses OverlayFs by default,
we need to change it to vfs. In order to change this configuration you need to edit a file called "deamon.json".
Once you have it installed, run the following command in order to create a local kubernetes cluster.

```bash
kind create cluster --name demo --config ./kind-config.yaml
```

### Setting up an NFS server on the cluster
An Valhalla Instance requires a volume that can be shared between multiple pods. In order to achieve that we are going to use an NFS server and a volume provisioner that will take advantage of this NFS server. Same as the previous step, you can skip this step if you already have a PersistentVolume provisioner installed on your cluster (if you are running in a cloud environment such as GCP for example).

First, we will install the NFS server:

```bash
kubectl create -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/deploy/example/nfs-provisioner/nfs-server.yaml
```

Next, we will install the provisioner using helm, so please make you have [helm](https://helm.sh/docs/intro/install/) installed on your machine.

```bash
helm repo add csi-driver-nfs https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/charts

helm install csi-driver-nfs csi-driver-nfs/csi-driver-nfs --namespace kube-system --version v3.1.0
```

Once you have a functional NFS server and an NFS volume provisioner you need create a new StorageClass on your cluster.

```bash
kubectl create -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/deploy/example/storageclass-nfs.yaml
```

## Installing the operator
Just run the following command to install the operator, includin the valhall CRD and all the other relevant resources.

```bash
kubectl apply -f https://github.com/itayankri/valhalla-operator/releases/latest/download/valhalla-operator.yaml
```

### Creating a new Valhalla Instance
This directory contains an example Valhalla resource. You can create it using kubectl.

```bash
kubectl apply -f example.yaml
```
