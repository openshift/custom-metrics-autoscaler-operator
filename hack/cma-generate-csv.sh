#!/usr/bin/env bash

if [ "$1" = "--dry-run" ]; then
  dry_run=1
  shift
fi

if ! [[ $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Usage: $0 [--dry-run] <version>"
  echo "Example: $0 2.7.1"
  echo "  would generate keda/2.7.1/manifests/cma.v2.7.1.clusterserviceversion.yaml"
  echo "  from keda/2.7.1/manifests/keda.v2.7.1.clusterserviceversion.yaml"
  echo "Example: $0 --dry-run 2.8.2"
  echo "  would generate keda/2.8.2/manifests/cma.v2.8.2.clusterserviceversion.yaml.new"
  echo "  from keda/2.8.2/manifests/keda.v2.8.2.clusterserviceversion.yaml"
  exit
fi

ver=$1

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd "$script_dir"/..

# This yaml data will overwrite fields from the upstream version. Any array fields must be removed first via the below jq filter
read -r -d '' cma_patch <<CMA_PATCH_EOF
metadata:
  annotations:
    description: Custom Metrics Autoscaler Operator, an event-driven autoscaler based upon KEDA
    operatorframework.io/suggested-namespace: openshift-keda
    operatorframework.io/cluster-monitoring: "true"
    operators.openshift.io/valid-subscription: '["OpenShift Kubernetes Engine", "OpenShift Container Platform", "OpenShift Platform Plus"]'
    repository: https://github.com/openshift/custom-metrics-autoscaler-operator
    support: Red Hat
    olm.skipRange: ">=2.7.1 <${ver}"
  name: custom-metrics-autoscaler.v${ver}
spec:
  description: "## About the managed application\\nCustom Metrics Autoscaler for OpenShift is an event driven autoscaler based upon KEDA.  Custom Metrics Autoscaler can monitor event sources like Kafka, RabbitMQ, or cloud event sources and feed the metrics from those sources into the Kubernetes horizontal pod autoscaler.  With Custom Metrics Autoscaler, you can have event driven and serverless scale of deployments within any Kubernetes cluster.\\n## About this Operator\\nThe Custom Metrics Autoscaler Operator deploys and manages installation of KEDA Controller in the cluster. Install this operator and follow installation instructions on how to install Custom Metrics Autoscaler in you cluster.\\n\\n## Prerequisites for enabling this Operator\\n## How to install Custom Metrics Autoscaler in the cluster\\nThe installation of Custom Metrics Autoscaler is triggered by the creation of \`KedaController\` resource. Please refer to the [KedaController Spec](https://github.com/openshift/custom-metrics-autoscaler-operator/blob/main/README.md#the-kedacontroller-custom-resource) for more deatils on available options.\\n\\nOnly resource named \`keda\` in namespace \`openshift-keda\` will trigger the installation, reconfiguration or removal of the KEDA Controller resource.\\n\\nThere should be only one KEDA Controller in the cluster. \\n"
  displayName: Custom Metrics Autoscaler
  icon:
  - base64data: PHN2ZyBpZD0iTGF5ZXJfMSIgZGF0YS1uYW1lPSJMYXllciAxIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxOTIgMTQ1Ij48ZGVmcz48c3R5bGU+LmNscy0xe2ZpbGw6I2UwMDt9PC9zdHlsZT48L2RlZnM+PHRpdGxlPlJlZEhhdC1Mb2dvLUhhdC1Db2xvcjwvdGl0bGU+PHBhdGggZD0iTTE1Ny43Nyw2Mi42MWExNCwxNCwwLDAsMSwuMzEsMy40MmMwLDE0Ljg4LTE4LjEsMTcuNDYtMzAuNjEsMTcuNDZDNzguODMsODMuNDksNDIuNTMsNTMuMjYsNDIuNTMsNDRhNi40Myw2LjQzLDAsMCwxLC4yMi0xLjk0bC0zLjY2LDkuMDZhMTguNDUsMTguNDUsMCwwLDAtMS41MSw3LjMzYzAsMTguMTEsNDEsNDUuNDgsODcuNzQsNDUuNDgsMjAuNjksMCwzNi40My03Ljc2LDM2LjQzLTIxLjc3LDAtMS4wOCwwLTEuOTQtMS43My0xMC4xM1oiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik0xMjcuNDcsODMuNDljMTIuNTEsMCwzMC42MS0yLjU4LDMwLjYxLTE3LjQ2YTE0LDE0LDAsMCwwLS4zMS0zLjQybC03LjQ1LTMyLjM2Yy0xLjcyLTcuMTItMy4yMy0xMC4zNS0xNS43My0xNi42QzEyNC44OSw4LjY5LDEwMy43Ni41LDk3LjUxLjUsOTEuNjkuNSw5MCw4LDgzLjA2LDhjLTYuNjgsMC0xMS42NC01LjYtMTcuODktNS42LTYsMC05LjkxLDQuMDktMTIuOTMsMTIuNSwwLDAtOC40MSwyMy43Mi05LjQ5LDI3LjE2QTYuNDMsNi40MywwLDAsMCw0Mi41Myw0NGMwLDkuMjIsMzYuMywzOS40NSw4NC45NCwzOS40NU0xNjAsNzIuMDdjMS43Myw4LjE5LDEuNzMsOS4wNSwxLjczLDEwLjEzLDAsMTQtMTUuNzQsMjEuNzctMzYuNDMsMjEuNzdDNzguNTQsMTA0LDM3LjU4LDc2LjYsMzcuNTgsNTguNDlhMTguNDUsMTguNDUsMCwwLDEsMS41MS03LjMzQzIyLjI3LDUyLC41LDU1LC41LDc0LjIyYzAsMzEuNDgsNzQuNTksNzAuMjgsMTMzLjY1LDcwLjI4LDQ1LjI4LDAsNTYuNy0yMC40OCw1Ni43LTM2LjY1LDAtMTIuNzItMTEtMjcuMTYtMzAuODMtMzUuNzgiLz48L3N2Zz4=
    mediatype: image/svg+xml
  install:
    spec:
      permissions:
      - rules:
        - apiGroups:
          - monitoring.coreos.com
          resources:
          - podmonitors
          verbs:
          - '*'
        serviceAccountName: custom-metrics-autoscaler-operator
  links:
  - name: Custom Metrics Autoscaler Documentation
    url: https://docs.openshift.com/container-platform/latest/nodes/pods/nodes-pods-autoscaling-custom.html
  maintainers:
  - email: support@redhat.com
    name: Red Hat
  provider:
    name: Red Hat
  version: ${ver}
CMA_PATCH_EOF

# build up jq_filter a little at a time, since it is very long

# delete array items from upstream so that we can populate them using the CMA data above
jq_filter='del(.spec.icon) | del(.spec.links) | del(.spec.maintainers) | '
# change all strings with value "keda-olm-operator" to "custom-metrics-autoscaler-operator"
jq_filter="$jq_filter"'walk(if type == "string" and . == "keda-olm-operator" then .="custom-metrics-autoscaler-operator" else . end) | '
# in CMA, we use olm.skipRange instead of replaces
jq_filter="$jq_filter"'del(.spec.replaces) | '
# update the json example CR so that it shows you how to install to "openshift-keda" namespace instead of "keda" namespace
jq_filter="$jq_filter"'.metadata.annotations."alm-examples" |= sub("\"namespace\": \"keda\""; "\"namespace\": \"openshift-keda\"") | '
# set the command to bash instead of /manager
jq_filter="$jq_filter"'.spec.install.spec.deployments[0].spec.template.spec.containers[0].command |= [ "/usr/bin/bash" ] |'
# export the env vars and then exec /manager
jq_filter="$jq_filter"'.spec.install.spec.deployments[0].spec.template.spec.containers[0].args |= ["-c", "export KEDA_OPERATOR_IMAGE=$RELATED_IMAGE_1; export KEDA_METRICS_SERVER_IMAGE=$RELATED_IMAGE_2; exec /manager \"$0\" \"$@\"" ] + .  |'
# add a port to the CMA operator so its metrics can be monitored
jq_filter="$jq_filter"'.spec.install.spec.deployments[0].spec.template.spec.containers[0].ports |= [{"containerPort":8080,"name":"http","protocol":"TCP"}] |'
# create a spot to pass in the operand image specs as env vars
jq_filter="$jq_filter"'.spec.install.spec.deployments[0].spec.template.spec.containers[0].env += [{"name":"RELATED_IMAGE_1","value":"CMA_OPERAND_PLACEHOLDER_1"},{"name":"RELATED_IMAGE_2","value":"CMA_OPERAND_PLACEHOLDER_2"}]'

# pipe the filtered upstream CSV and the patch together to jq to combine them
{ bin/yaml2json keda/${ver}/manifests/keda.v${ver}.clusterserviceversion.yaml | jq "$jq_filter";
  echo "$cma_patch" | bin/yaml2json; } | \
  jq --slurp 'reduce .[] as $item ({}; . * $item)' | bin/json2yaml > keda/${ver}/manifests/cma.v${ver}.clusterserviceversion.yaml.new

if ! test "$dry_run" = "1"; then
  mv keda/${ver}/manifests/cma.v${ver}.clusterserviceversion.yaml.new keda/${ver}/manifests/cma.v${ver}.clusterserviceversion.yaml
fi
