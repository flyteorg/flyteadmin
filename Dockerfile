# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

FROM golang:1.17.1-alpine3.14 as builder
RUN apk add git openssh-client make curl

# COPY only the go mod files for efficient caching
COPY go.mod go.sum /go/src/github.com/flyteorg/flyteadmin/
WORKDIR /go/src/github.com/flyteorg/flyteadmin

# Pull dependencies
RUN go mod download

# COPY the rest of the source code
COPY . /go/src/github.com/flyteorg/flyteadmin/

# This 'linux_compile' target should compile binaries to the /artifacts directory
# The main entrypoint should be compiled to /artifacts/flyteadmin
RUN make linux_compile

# update the PATH to include the /artifacts directory
ENV PATH="/artifacts:${PATH}"

# This will eventually move to centurylink/ca-certs:latest for minimum possible image size
FROM alpine:3.14
LABEL org.opencontainers.image.source https://github.com/flyteorg/flyteadmin

COPY --from=builder /artifacts /bin

# Ensure the latest CA certs are present to authenticate SSL connections.
RUN apk --update add ca-certificates

# Set the uid,gid,group and user
ENV UID=1001
ENV GID=2001
ENV GROUP=flyte
ENV USER=flyte
# Add flyte user and group
RUN addgroup -g ${GID} ${GROUP} 
RUN adduser -D -s /bin/sh -G ${GROUP} -u ${UID} ${USER}

CMD ["flyteadmin"]
