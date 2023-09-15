# Valhalla Kubernetes Operator
A kubernetes operator to deploy and manage [Valhalla](https://valhalla.readthedocs.io/en/latest/valhalla-intro/) routing engine. This operator efficiently deploys Valhalla instances by sharing map data accross all pods of a specific instance.

## Quickstart
First, make sure you have a running Kubernetes cluster and kubectl installed to access it. Then run the following command to install the operator:
```
kubectl apply -f https://github.com/itayankri/valhalla-operator/releases/latest/download/valhalla-operator.yaml
```

Then you can deploy a Valhalla instance:
```
kubectl apply -f https://github.com/itayankri/valhalla-operator/blob/master/examples/example.yaml
```
For a full setup from scratch checkout this [Medium]([https://github.com/itayankri/valhalla-operator/tree/master/examples](https://medium.com/@itay.ankri/deploying-valhalla-routing-engine-on-kubernetes-using-valhalla-operator-2426e79ac746)).

## Pausing the Operator
The reconciliation can be paused by adding the following annotation to the Valhalla resource:
```bash
valhalla.itayankri/operator.paused: "true"
```
The operator will not react to any changes to the Valhalla resource or any of the watched resources. If a paused Valhalla resource is deleted, the dependent resources will still be cleaned up because thay all have an ownerReference.
