# zarf-xsoar-package
This package deploys Palo Alto XSOAR into an environment.

Note: Prior to running zarf deploy, a manual upload of the liscence file is required. To do this, take your liscense file, and run the following command:

kubectl create configmap xsoar-lic --from-file=<liscense_file> -n xsoar