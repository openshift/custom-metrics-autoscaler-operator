#!/usr/bin/env bash

# we only care about CSV files after 2.7.1
csv_versions="$(ls keda | sort --version-sort | sed -n '/^2\.7\.1$/,$ p')"

script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

cd "$script_dir"/..

rv=0

for ver in $csv_versions; do
  f="keda/${ver}/manifests/cma.v${ver}.clusterserviceversion.yaml"
  if ! test -f "$f"; then
    echo "Custom Metrics Autoscaler CSV file missing: $f"
    echo "To create it, run hack/cma-generate-csv.sh $ver"
    rv=1
    continue
  fi
  "$script_dir"/cma-generate-csv.sh --dry-run $ver
  if ! diff -q "$f".new "$f"; then
    echo "Unexpected differences found in Custom Metrics Autoscaler CSV file: $f"
    diff -u "$f".new "$f"
    echo "To correct differences, run hack/cma-generate-csv.sh $ver"
    rv=1
  fi
  rm -f "$f".new
done

exit $rv
