# Portworx Manifests

This repository contains Kubernetes (K8s) manifests needed for deploying the Portworx Operator and the Portworx StorageCluster Custom Resource Definition (CRD). These manifests are stored in the `storage` folder.

## Purpose

The `storage` folder serves as a collection of YAML files, or 'manifests', which are used to create, configure and manage Kubernetes resources for deploying the Portworx Operator and the StorageCluster CRD. 

Portworx is a cloud-native storage platform that provides high availability, data protection, data security, and data mobility. The Portworx Operator manages the lifecycle of Portworx, and the StorageCluster CRD represents a Portworx Cluster.

The manifests are named in the numerical order they need to be deployed. This order is important and should be followed to ensure a correct setup.

## Future Work

This is an initial setup for the Portworx deployment. As the project progresses, these manifests will be refined, tested, and tweaked to implement best security practices. 

At the moment, we are using a basic configuration to establish a baseline functionality and ensure everything is working as expected within zarf

Please follow this repository to keep up with the changes and improvements over time.
