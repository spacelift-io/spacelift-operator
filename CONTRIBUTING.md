# Contributor guide

## How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

## Run the controller locally

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Create the K8s cluster

```sh
kind create cluster
```

If it already exists, you need to delete it first:

```sh
kubectl delete cluster kind
```

### Install prerequisites

```sh
make kustomize
make controller-gen
make envtest
```

### Create the `spacelift-credentials` secret

To authenticate with the Spacelift API, you need to have an [API Key](https://docs.spacelift.io/integrations/api#spacelift-api-key-token) created in Spacelift.

After that, you need to create a secret in your Kubernetes cluster called `spacelift-credentials` with the following keys:

- `SPACELIFT_API_KEY_ENDPOINT` - the endpoint of the Spacelift API (for example `https://<account-name>.app.spacelift.io`)
- `SPACELIFT_API_KEY_ID` - the ID of the API key
- `SPACELIFT_API_KEY_SECRET` - the secret of the API key

An example of how to create the secret:

```sh
kubectl create secret generic spacelift-credentials --from-literal=SPACELIFT_API_KEY_ENDPOINT='https://mycorp.app.spacelift.io' --from-literal=SPACELIFT_API_KEY_ID='01HV1GND58KS3MFNWM5BLF33D' --from-literal=SPACELIFT_API_KEY_SECRET='3cbef141b857f40042351c79d6d435b6c1e277662ac828ef3b6cf'
```

### Install the CRDs into the cluster

```sh
make install
```

### Run the controller

There is a pre-configured configuration for VS Code in `.vscode/launch.json` that you can use to run the controller in debug mode, but it's basically just simply starting `./cmd/main.go`. A shortcut is `make run`.

Once the operator is up and running, you can create an example Space.

### Create an example Space

Create a yaml file in a `.gitignore`d directory (e.g. `bin/` or `dist/`) with the following content:

```yaml
apiVersion: app.spacelift.io/v1beta1
kind: Space
metadata:
  name: space-test
spec:
  name: created-from-operator
  parentSpace: root
  inheritEntities: true
  description: "This space was created from the K8s operator"
```

Then apply it to the cluster:

```sh
kubectl apply -f bin/space.yaml
```

To delete the Space, run:

```sh
kubectl delete space space-test
```

Note that this command will **not** delete the Space from Spacelift, but only from the Kubernetes cluster. As of now, the operator does not support removing resources from Spacelift.

## Day to day development

### Modifying the API definitions

If you are editing the API definitions (`api/v1beta1/` folder), you need to regenerate the manifests:

```sh
make controller-manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
