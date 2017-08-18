#!/bin/bash
set -x

kubectl delete deployment -l app=krank -n kube-system
kubectl delete service -l app=krank -n kube-system

# Delete RBAC objects, if --rbac flag was used.
kubectl delete serviceaccount -l app=krank -n kube-system
kubectl delete clusterrolebindings -l app=krank -n kube-system
kubectl delete clusterrole -l app=krank -n kube-system
