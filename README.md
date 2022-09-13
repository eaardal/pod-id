# Pod ID

Prints the full ID of a Kubernetes pod by a partial name

## Motivation

Pods in kubernetes usually has names such as
```
my-app-123sdk-54lda5
other-app-789sdk-89jds
```
To get logs, the workflow is to first use `kubectl get pod` to list them all, then copy the ID of the pod you want,
then run `kubectl logs <pod id>` to get the logs.

This app aims to remove the list-pods-and-copy-id steps. With pod-id you can instead run `kubectl logs $(podid my-app)`.

Pod-id will use your current Kubernetes context in `~/.kube/config` to read all pods and find the ones matching the
app name you provide as first argument. It is recommended you use [kubens](https://github.com/ahmetb/kubectx) to manage your current namespace (see usage).

## Install

Build it:
```shell
go build -o podid main.go
```

Put `podid` somewhere on your PATH.

## Usage

Standalone:
```shell
podid my-app # my-app-123sdk-54lda5
```

With kubectl:
```
kubectl logs $(podid my-app)
```

App names can be partial. If you have a pod with ID `card-payment-api-123sdk-54lda5` you can use a part of the name such as `podid payment` to look it up.

### Resolving kubernetes namespace

TL;DR: Use [kubens](https://github.com/ahmetb/kubectx). Pod-id will read `kubens --current` to find the current namespace.

Internally, pod-id needs to know in which kubernetes namespace to fetch pods. This can be done in a few ways:

First it will check if you gave namespace as part of the app name. Some examples:
```shell
podid <namespace>/<app>
podid monitoring/my-app # namespace=monitoring, appName=my-app
podid apps/other-app # namespace=apps, appName=other-app
```

If no namespace was found that way, it will check if the environment variable `PODID_NAMESPACE` is set and use that if it is.

If not, it will try the last method of executing the shell command `kubens --current` and reading the result.

Therefore, the most reliable and hassle-free method is using [kubens](https://github.com/ahmetb/kubectx) to manage your active kubernetes context and namespace.

### Options

#### Pod number

If an app has multiple pods such as:
```shell
my-app-123sdk-54lda5
my-app-465dsa-sd87d8
```
Then adding the pod number after the app name will print the ID for that number:
```shell
podid my-app 1 # my-app-123sdk-54lda5
podid my-app 2 # my-app-465dsa-sd87d8
```

#### Clipboard

Also add the "copy" parameter to also copy the pod id to your clipboard:
```shell
podid my-app copy # Clipboard is now: my-app-123sdk-54lda5
podid my-app 2 copy # Clipboard is now: my-app-465dsa-sd87d8
```
It will still print the pod id to stdout, so you can still use it like `kubectl logs $(podid my-app copy)` and get the logs in addition to having the ID copied to your clipboard. 