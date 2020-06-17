#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

echo "Loading github docker images into 'kind' cluster to workaround this issue: https://github.com/containerd/containerd/issues/3291#issuecomment-631746985"
docker login --username ${DOCKER_USERNAME} --password ${DOCKER_PASSWORD} docker.pkg.github.com

docker tag "my_awesome_image:latest" "flyteadmin:test"
kind load docker-image flyteadmin:test

# start flyteadmin and dependencies
kubectl apply -f "${DIR}/k8s/integration.yaml"

# wait for flyteadmin deployment to complete
kubectl -n flyte rollout status deployment flyteadmin

# get the name of the flyteadmin pod
POD_NAME=$(kubectl get pods -n flyte -o go-template="{{range .items}}{{.metadata.name}}:{{end}}" | tr ":" "\n" | grep flyteadmin | grep Running)

# launch the integration tests
kubectl exec -it -n flyte "$POD_NAME" -- make -C /go/src/github.com/lyft/flyteadmin integration
