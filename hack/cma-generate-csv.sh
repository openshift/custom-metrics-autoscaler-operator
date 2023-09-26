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
# See https://docs.engineering.redhat.com/display/CFC/Best_Practices#Best_Practices-(New)RequiredInfrastructureAnnotations for
# information on features.oeprators.openshift.io/* annotations
read -r -d '' cma_patch <<CMA_PATCH_EOF
metadata:
  annotations:
    description: Custom Metrics Autoscaler Operator, an event-driven autoscaler based upon KEDA
    features.operators.openshift.io/disconnected: "true"
    features.operators.openshift.io/fips-compliant: "false"
    features.operators.openshift.io/proxy-aware: "false"
    features.operators.openshift.io/cnf: "false"
    features.operators.openshift.io/cni: "false"
    features.operators.openshift.io/csi: "false"
    features.operators.openshift.io/tls-profiles: "false"
    features.operators.openshift.io/token-auth-aws: "false"
    features.operators.openshift.io/token-auth-azure: "false"
    features.operators.openshift.io/token-auth-gcp: "false"
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
  - base64data: PHN2ZyBlbmFibGUtYmFja2dyb3VuZD0ibmV3IDAgMCAzOCAzOCIgdmlld0JveD0iMCAwIDM4IDM4IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Im0yNy43IDEuNmgtMTcuNGMtNC44IDAtOC43IDMuOS04LjcgOC43djE3LjRjMCA0LjggMy45IDguNyA4LjcgOC43aDE3LjRjNC44IDAgOC43LTMuOSA4LjctOC43di0xNy40YzAtNC44LTMuOS04LjctOC43LTguN3oiIGZpbGw9IiNmZmYiLz48cGF0aCBkPSJtMjggMi4yYzQuMyAwIDcuOCAzLjUgNy44IDcuOHYxOGMwIDQuMy0zLjUgNy44LTcuOCA3LjhoLTE4Yy00LjMgMC03LjgtMy41LTcuOC03Ljh2LTE4YzAtNC4zIDMuNS03LjggNy44LTcuOHptMC0xLjJoLTE4Yy01IDAtOSA0LTkgOXYxOGMwIDUgNCA5IDkgOWgxOGM1IDAgOS00IDktOXYtMThjMC01LTQtOS05LTl6Ii8+PHBhdGggZD0ibTI4IDI0LjFjLS4yIDAtLjQtLjEtLjUtLjMtLjItLjMtLjEtLjcuMi0uOWwuNi0uM3YtLjZjMC0uMy4zLS42LjYtLjZzLjYuMy42LjZ2MWMwIC4yLS4xLjQtLjMuNWwtLjkuNWMtLjEuMS0uMi4xLS4zLjF6Ii8+PHBhdGggZD0ibTI4LjkgMjAuMmMtLjMgMC0uNi0uMy0uNi0uNnYtMS4yYzAtLjMuMy0uNi42LS42cy42LjMuNi42djEuMmMwIC4zLS4zLjYtLjYuNnoiLz48cGF0aCBkPSJtMjguOSAxNi42Yy0uMyAwLS42LS4zLS42LS42di0uNmwtLjYtLjRjLS4zLS4yLS40LS42LS4yLS45cy42LS40LjktLjJsLjkuNWMuMi4xLjMuMy4zLjV2MWMtLjEuNC0uNC43LS43Ljd6Ii8+PHBhdGggZD0ibTI1LjkgMTMuOWMtLjEgMC0uMiAwLS4zLS4xbC0xLS42Yy0uMy0uMi0uNC0uNi0uMi0uOXMuNi0uNC45LS4ybDEgLjZjLjMuMi40LjYuMi45LS4yLjItLjQuMy0uNi4zeiIvPjxwYXRoIGQ9Im0yMi44IDEyLjFjLS4xIDAtLjIgMC0uMy0uMWwtLjYtLjMtLjUuM2MtLjMuMi0uNy4xLS45LS4ycy0uMS0uNy4yLS45bC45LS41Yy4yLS4xLjQtLjEuNiAwbC45LjVjLjMuMi40LjYuMi45LS4xLjItLjMuMy0uNS4zeiIvPjxwYXRoIGQ9Im0yMS45IDI3LjZjLS4xIDAtLjIgMC0uMy0uMWwtLjgtLjVjLS4zLS4yLS40LS42LS4yLS45cy42LS40LjktLjJsLjYuMy42LS4zYy4zLS4yLjctLjEuOS4ycy4xLjctLjIuOWwtLjkuNWMtLjQuMS0uNS4xLS42LjF6Ii8+PHBhdGggZD0ibTI0LjkgMjUuOWMtLjIgMC0uNC0uMS0uNS0uMy0uMi0uMy0uMS0uNy4yLS45bDEtLjZjLjMtLjIuNy0uMS45LjJzLjEuNy0uMi45bC0xIC42Yy0uMi4xLS4zLjEtLjQuMXoiLz48cGF0aCBkPSJtMjMgMjN2LThsLTYuOS00LTcgNHY4bDcgNHoiIGZpbGw9IiNmZmYiLz48cGF0aCBkPSJtMTYuMSAyNy42Yy0uMSAwLS4yIDAtLjMtLjFsLTYuOS00Yy0uMi0uMS0uMy0uMy0uMy0uNXYtOGMwLS4yLjEtLjQuMy0uNWw2LjktNGMuMi0uMS40LS4xLjYgMGw2LjkgNGMuMi4xLjMuMy4zLjV2OGMwIC4yLS4xLjQtLjMuNWwtNi45IDRjLS4xLjEtLjIuMS0uMy4xem0tNi4zLTUgNi4zIDMuNiA2LjMtMy42di03LjNsLTYuMy0zLjYtNi4zIDMuNnoiLz48cGF0aCBkPSJtMTguNiAxNy45Yy0uMS0uMi0uMy0uMy0uNS0uM2gtMXYtMy4xYzAtLjMtLjItLjUtLjUtLjZzLS42LjEtLjcuNGwtMi4zIDUuM2MtLjEuMi0uMS40IDAgLjZzLjMuMy41LjNoMXYzLjFjMCAuMy4yLjUuNS42aC4xYy4yIDAgLjUtLjEuNi0uNGwyLjMtNS4yYy4xLS4zLjEtLjUgMC0uN3oiIGZpbGw9IiNlMDAiLz48L3N2Zz4=
    mediatype: image/svg+xml
  links:
  - name: Custom Metrics Autoscaler Documentation
    url: https://docs.openshift.com/container-platform/latest/nodes/cma/nodes-cma-autoscaling-custom.html
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
# change keda -> CMA in replaces
jq_filter="$jq_filter"'.spec.replaces |= sub("keda.v";"custom-metrics-autoscaler.v") | '
# update the json example CR so that it shows you how to install to "openshift-keda" namespace instead of "keda" namespace
jq_filter="$jq_filter"'.metadata.annotations."alm-examples" |= sub("\"namespace\": \"keda\""; "\"namespace\": \"openshift-keda\"") | '
# set the command to bash instead of /manager
jq_filter="$jq_filter"'.spec.install.spec.deployments[0].spec.template.spec.containers[0].command |= [ "/usr/bin/bash" ] |'
# export the env vars and then exec /manager. This hack is needed due to OSBS requiring that env vars have the prefix RELATED_IMAGE_, so bash translates from OSBS-style env vars to operator env vars. See https://osbs.readthedocs.io/en/latest/users.html?highlight=operator#pullspec-locations
jq_filter="$jq_filter"'.spec.install.spec.deployments[0].spec.template.spec.containers[0].args |= ["-c", "export KEDA_OPERATOR_IMAGE=$RELATED_IMAGE_1; export KEDA_METRICS_SERVER_IMAGE=$RELATED_IMAGE_2; export KEDA_ADMISSION_WEBHOOKS_IMAGE=$RELATED_IMAGE_3; exec /manager \"$0\" \"$@\"" ] + .  |'
# create a spot to pass in the operand image specs as env vars using OSBS-style RELATED_IMAGE_ names
jq_filter="$jq_filter"'.spec.install.spec.deployments[0].spec.template.spec.containers[0].env += [{"name":"RELATED_IMAGE_1","value":"CMA_OPERAND_PLACEHOLDER_1"},{"name":"RELATED_IMAGE_2","value":"CMA_OPERAND_PLACEHOLDER_2"},{"name":"RELATED_IMAGE_3","value":"CMA_OPERAND_PLACEHOLDER_3"}]'

# pipe the filtered upstream CSV and the patch together to jq to combine them
{ bin/yaml2json keda/${ver}/manifests/keda.v${ver}.clusterserviceversion.yaml | jq "$jq_filter";
  echo "$cma_patch" | bin/yaml2json; } | \
  jq --slurp 'reduce .[] as $item ({}; . * $item)' | bin/json2yaml > keda/${ver}/manifests/cma.v${ver}.clusterserviceversion.yaml.new

if ! test "$dry_run" = "1"; then
  mv keda/${ver}/manifests/cma.v${ver}.clusterserviceversion.yaml.new keda/${ver}/manifests/cma.v${ver}.clusterserviceversion.yaml
fi
