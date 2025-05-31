# ctf-ui-operator

ctf-ui-operator is a Kubernetes operator designed to automate the deployment and management of Capture The Flag (CTF)
web UIs within a Kubernetes cluster. It is intended to be used together with the
[ctf-challenge-operator](https://github.com/Backbone81/ctf-challenge-operator) to automatically configure the CTF
challenges for the web UI.

It currently provisions [CTFd](https://github.com/CTFd/CTFd) with all its dependencies (Redis, MariaDB, Minio) providing
a quick and easy solution for running a CTF event.

**NOTE: This project is currently in early development and is not yet in a state to be actually used in a CTF event
or even in a proof-of-concept situation.**

## Description

This project provides one Kubernetes custom resource definition to help with running a CTF event:

- `CTFd`: This resource describes a single CTFd instance and its initial configuration.

**NOTE: There are other CRDs like `Redis`, `MariaDB` or `Minio` which are dependencies for `CTFd`. Those are not
intended to be used directly.**

## Getting Started

To deploy this operator into your Kubernetes cluster:

```shell
kubectl apply -k https://github.com/backbone81/ctf-ui-operator/manifests?ref=v0.1.0
```

To deploy a new instance of `CTFd` create a yaml manifest `ctfd-sample.yaml`:

```yaml
---
apiVersion: ui.ctf.backbone81/v1alpha1
kind: CTFd
metadata:
  name: ctfd-sample
spec:
  title: Demo CTF
  description: This is a demo CTF.
```

Apply this manifest to your cluster:

```shell
kubectl apply -f ctfd-sample.yaml
```

The first startup might need a few minutes to pull all the necessary docker images.

Port-forward the service of your instance:

```shell
kubectl port-forward svc/ctfd-sample 3000:http
```

Log into your instance with the admin credentials stored in the secret `ctfd-sample-admin`.

For a production deployment, you probably want to provide an Ingress for your instance, which is out of scope of this
operator. You also might want to tweak the settings of your instance. See `examples/crd-sample.yaml` for a more
elaborate setup or `api/v1alpha1/ctfd.go` for details on all the available settings.

### Operator Command Line Parameters

The operator provides the following command line parameters:

```text
This operator manages CTF UI instances.

Usage:
  ctf-ui-operator [flags]

Flags:
      --enable-developer-mode              This option makes the log output friendlier to humans.
      --health-probe-bind-address string   The address the probe endpoint binds to. (default "0")
  -h, --help                               help for ctf-ui-operator
      --kubernetes-client-burst int        The number of burst queries the Kubernetes client is allowed to send against the Kubernetes API. (default 10)
      --kubernetes-client-qps float32      The number of queries per second the Kubernetes client is allowed to send against the Kubernetes API. (default 5)
      --leader-election-enabled            Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
      --leader-election-id string          The ID to use for leader election. (default "ctf-ui-operator")
      --leader-election-namespace string   The namespace in which leader election should happen. (default "ctf-ui-operator")
      --log-level int                      How verbose the logs are. Level 0 will show info, warning and error. Level 1 and up will show increasing details.
      --metrics-bind-address string        The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service. (default "0")
```

## Development

This project intends to be run on cloud provider infrastructure. As cloud providers provide new Kubernetes version only
after some time, this project aims at the oldest supported Kubernetes version. See
[Supported versions](https://kubernetes.io/releases/version-skew-policy/#supported-versions) of the official Kubernetes
documentation.

This project uses tools like `controller-gen` which was sensitive to Go version updates in the past. To reduce the
likelyhood of Go versions breaking the toolchain, this project aims at the oldest supported Go version. See
[Release Policy](https://go.dev/doc/devel/release#policy) of the official Go documentation.

You need to have the following tools available

- Go
- Docker
- make
- Linux

**NOTE: Windows is currently not supported. MacOS might work but is untested.**

For setting up your local Kubernetes development cluster:

```shell
make init-local
```

This will create a Kubernetes cluster with kind and install the CRDs.

To run the code:

```shell
make run
```

To run the tests:

```shell
make test
```

To run the end-to-end tests:

```shell
make test-e2e
```

If you changed the data types of the custom resources, you can install the updated version with:

```shell
make install
```

To clean up everything (including the kind cluster you created with `make init-local`):

```shell
make clean
```

### Third Party Tools

All third party tools required for development are provided through shims in the `bin` directory of this project. Those
shims are shell scripts which download the required tool on-demand into the `tmp` directory and forward any arguments
to the real executable. If you want to interact with those tools outside of makefile targets, add the `bin` directory to
your `PATH` environment variable like this:

```shell
export PATH=${PWD}/bin:${PATH}
```

### Building a Release

To build a new release:

- Pick the next version to use as a git tag and a docker image tag. This should be `v` followed by a semantic version.
  Let's assume `v1.2.3` as an example for the new version.
- Update the docker image in the manifests subdirectory to the new docker image tag for the version. That would be
  `backbone81/ctf-ui-operator:v1.2.3`.
- Update the git tag in the installation section of the README.md to the new release.
- Clean up your local development environment and run the tests and end-to-end tests locally:
  ```shell
  make clean
  make test
  make test-e2e
  ```
  If anything fails, fix the errors.
- Commit and push your changes. Wait for the pipeline to succeed. If the pipeline fails, fix the errors.
- Create a git tag for the release and push the tag:
  ```shell
  git tag v1.2.3
  git push origin v1.2.3
  ```
- Wait for the pipeline to succeed and publish the new docker image. If the pipeline fails, fix the errors and create
  a new release. Do not delete the old release.
