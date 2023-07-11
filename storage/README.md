Portworx Storage Cluster Deployment
This repository includes all the necessary configurations to deploy a Portworx storage cluster on a Kubernetes environment.

Quick Start
You can quickly deploy everything you need for the Portworx storage cluster by running the provided shell script. The script will execute all necessary YAML files in the appropriate order, with a one-minute pause between each file to ensure proper resource allocation and initialization.

Running the Deployment Script
Use the following command to execute the script:

bash <(curl -s https://raw.githubusercontent.com/naps-dev/dco-core/main/storage/manifests/t1-portworx.sh)

The script will:

Apply the 01-px-operator.yaml file, deploying the Portworx operator.
Pause for 1 minute to allow the operator to initialize.
Apply the 02-px-stc.yaml file, deploying the Portworx storage cluster.
Pause for 1 minute to allow the storage cluster to initialize.
Apply the 03-px-sc.yaml file, setting up the default Portworx storage class.
The total execution time is approximately 8 minutes. At the end of the process, your Portworx storage cluster and the default storage class px should be fully deployed and operational.


For more detailed information or manual deployment steps, please refer to the respective YAML files.

This README provides a brief explanation of the deployment process and the functionality of the deployment script. You can expand the README to include more detailed information about your project and the Portworx storage cluster.
