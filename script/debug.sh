dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ${GOPATH}/bin/flyteadmin serve -- --config  flyteadmin_config.yaml --server.kube-config ~/.kube/config

