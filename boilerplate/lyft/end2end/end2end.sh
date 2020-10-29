#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

OUT="${DIR}/tmp"
rm -rf ${OUT}
git clone https://github.com/lyft/flyte.git "${OUT}"

pushd ${OUT}

# TODO: load all images
if [ ! -z "$PROPELLER" ]; then
  kind load docker-image ${PROPELLER}
  sed -i.bak -e "s_flytepropeller:.*_${PROPELLER}_g" ${OUT}/kustomize/base/propeller/deployment.yaml
fi

if [ ! -z "$ADMIN" ]; then
  kind load docker-image ${ADMIN}
  sed -i.bak -e "s_flyteadmin:.*_${ADMIN}_g" ${OUT}/kustomize/base/admindeployment/deployment.yaml
fi

make kustomize
make end2end_execute
popd
