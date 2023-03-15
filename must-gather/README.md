# OpenShift Custom Metric Autoscaler Operator Must-Gather
`custom-metric-autoscaler-must-gather` is a tool built on top of [OpenShift must-gather](https://github.com/openshift/must-gather)
that expands its capabilities to gather specific information for the OpenShift Custom Metric Autoscaler Operator.

## Usage
To gather only Openshift Custom Metric Autoscaler Operator information use the following command: 
```sh
  oc adm must-gather --image="$(oc get packagemanifests openshift-custom-metrics-autoscaler-operator \
    -n openshift-marketplace \
    -o jsonpath='{.status.channels[?(@.name=="stable")].currentCSVDesc. annotations.containerImage}')"
```
where the custom image for the must-gather command is pulled directly from the operator' package manifests, so that 
it works on any cluster with Custom Metric Autoscaler Operator is available.

To gather default [OpenShift must-gather](https://github.com/openshift/must-gather) in addition to Openshift Custom
Metric Autoscaler Operator information you should fetch the operator image and can combine both images with the 
following command:
```sh
# fetch operator image
IMAGE="$(oc get packagemanifests openshift-custom-metrics-autoscaler-operator \
  -n openshift-marketplace \
  -o jsonpath='{.status.channels[?(@.name=="stable")].currentCSVDesc.annotations.containerImage}')"

# invoke must-gather with addition image
oc adm must-gather --image-stream=openshift/must-gather --image=${IMAGE}
```

## Collection script data
As a result of the above commands a local directory will be crated with a dump of the resources for Openshift Custom
Metric Autoscaler Operator

You will get a dump of:
- The `openshift-keda` namespace and its children objects
- The custom-metric-autoscaler operator install objects
- All custom-metric-autoscaler CRD's definitions, if present. That is:
  - kedacontroller
  - scaledjob
  - scaledobject
  - triggerauthentication
  - clustertriggerauthentication

In order to get data about other parts of the cluster that are not specific to custom-metric-autoscaler operator
you should run `oc adm must-gather` without passing the custom image. Run `oc adm must-gather -h` to see more options.

## Must gather output
Example must-gather output for `custom-metric-autoscaler-must-gather` tool:
```
└── openshift-keda
    ├── apps
    │   ├── daemonsets.yaml
    │   ├── deployments.yaml
    │   ├── replicasets.yaml
    │   └── statefulsets.yaml
    ├── apps.openshift.io
    │   └── deploymentconfigs.yaml
    ├── autoscaling
    │   └── horizontalpodautoscalers.yaml
    ├── batch
    │   ├── cronjobs.yaml
    │   └── jobs.yaml
    ├── build.openshift.io
    │   ├── buildconfigs.yaml
    │   └── builds.yaml
    ├── core
    │   ├── configmaps.yaml
    │   ├── endpoints.yaml
    │   ├── events.yaml
    │   ├── persistentvolumeclaims.yaml
    │   ├── pods.yaml
    │   ├── replicationcontrollers.yaml
    │   ├── secrets.yaml
    │   └── services.yaml
    ├── discovery.k8s.io
    │   └── endpointslices.yaml
    ├── image.openshift.io
    │   └── imagestreams.yaml
    ├── k8s.ovn.org
    │   ├── egressfirewalls.yaml
    │   └── egressqoses.yaml
    ├── keda.sh
    │   ├── kedacontrollers
    │   │   └── keda.yaml
    │   ├── scaledobjects
    │   │   └── example-scaledobject.yaml
    │   └── triggerauthentications
    │       └── example-triggerauthentication.yaml
    ├── monitoring.coreos.com
    │   └── servicemonitors.yaml
    ├── networking.k8s.io
    │   └── networkpolicies.yaml
    ├── openshift-keda.yaml
    ├── pods
    │   ├── custom-metrics-autoscaler-operator-58bd9f458-ptgwx
    │   │   ├── custom-metrics-autoscaler-operator
    │   │   │   └── custom-metrics-autoscaler-operator
    │   │   │       └── logs
    │   │   │           ├── current.log
    │   │   │           ├── previous.insecure.log
    │   │   │           └── previous.log
    │   │   └── custom-metrics-autoscaler-operator-58bd9f458-ptgwx.yaml
    │   ├── custom-metrics-autoscaler-operator-58bd9f458-thbsh
    │   │   └── custom-metrics-autoscaler-operator
    │   │       └── custom-metrics-autoscaler-operator
    │   │           └── logs
    │   ├── keda-metrics-apiserver-65c7cc44fd-6wq4g
    │   │   ├── keda-metrics-apiserver
    │   │   │   └── keda-metrics-apiserver
    │   │   │       └── logs
    │   │   │           ├── current.log
    │   │   │           ├── previous.insecure.log
    │   │   │           └── previous.log
    │   │   └── keda-metrics-apiserver-65c7cc44fd-6wq4g.yaml
    │   └── keda-operator-776cbb6768-fb6m5
    │       ├── keda-operator
    │       │   └── keda-operator
    │       │       └── logs
    │       │           ├── current.log
    │       │           ├── previous.insecure.log
    │       │           └── previous.log
    │       └── keda-operator-776cbb6768-fb6m5.yaml
    ├── policy
    │   └── poddisruptionbudgets.yaml
    └── route.openshift.io
        └── routes.yaml
```
