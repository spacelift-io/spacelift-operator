# Spacelift Operator

This repository implements a Kubernetes Operator for managing Spacelift resources in a Kubernetes cluster.

## Description

The Controller is responsible for creating and updating resources in Spacelift based on custom resource definitions (CRD). The following resources are supported:

- Stacks
- Runs
- Spaces
- Contexts
- Policies

⚠️ Note: Currently we do not delete resources in Spacelift when the corresponding custom resource is deleted.

## Installing

To install the Spacelift Operator along with its CRDs, run the following command:

```sh
kubectl apply -f https://downloads.spacelift.io/spacelift-operator/latest/manifests.yaml
```

> You can download the manifests yourself from <https://downloads.spacelift.io/spacelift-operator/latest/manifests.yaml> if you would like to inspect them or alter the Deployment configuration for the controller.

### Create the `spacelift-credentials` secret

To authenticate with the Spacelift API, you need to have an [API Key](https://docs.spacelift.io/integrations/api#spacelift-api-key-token) created in Spacelift.

After that, you need to create a secret in your Kubernetes cluster called `spacelift-credentials` with the following keys:

- `SPACELIFT_API_KEY_ENDPOINT` - the endpoint of the Spacelift API (`https://<account-name>.app.spacelift.io`)
- `SPACELIFT_API_KEY_ID` - the ID of the API key
- `SPACELIFT_API_KEY_SECRET` - the secret of the API key

An example of how to create the secret:

```sh
kubectl create secret generic spacelift-credentials --from-literal=SPACELIFT_API_KEY_ENDPOINT='https://mycorp.app.spacelift.io' --from-literal=SPACELIFT_API_KEY_ID='01HV1GND58KS3MFNWM5BLF33D' --from-literal=SPACELIFT_API_KEY_SECRET='3cbef141b857f40042351c79d6d435b6c1e277662ac828ef3b6cf'
```

### Create a Spacelift resource

You can now create a Spacelift resource in your Kubernetes cluster. For example, to create a Spacelift Stack, you can use the following manifest:

```sh
kubectl apply -f - <<EOF
apiVersion: app.spacelift.io/v1beta1
kind: Stack
metadata:
  name: stack-test
spec:
  name: spacelift-operator-test
  settings:
    spaceName: root
    repository: organization/repo-name
    branch: main
    vendorConfig:
      terraform:
        version: 1.7.1
        workflowTool: OPEN_TOFU
EOF
```

## Contributing and local setup

If you need to make change to this project, please read the [CONTRIBUTING.md](./CONTRIBUTING.md) file carefully.

## Releasing

To release a new version of the operator, just tag the repo with the latest version number:

```shell
git checkout main && git pull
git tag v<version>
git push origin v<version>
```

This will trigger the release workflow, and will push the latest container image to ECR, upload the Kubernetes manifests, and create a GitHub release.
