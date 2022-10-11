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