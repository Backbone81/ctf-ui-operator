# ctf-ui-operator

TODO

**NOTE: This project is currently in early development and is not yet in a state to be actually used in a CTF event
or even in a proof-of-concept situation.**

## Description

TODO

## Getting Started

To deploy this operator into your Kubernetes cluster:

```shell
kubectl apply -k https://github.com/backbone81/ctf-ui-operator/manifests?ref=main
```

**NOTE: As there is not yet a real release of this operator, the docker image referenced in that manifest does not
exist yet.**

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
