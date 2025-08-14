# Timeprinter controller

## Build

```sh
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

go generate ./...
```


## Use it


Deploy the CRDs

```sh
kubectl apply -f config/crd/bases
```

Create a TimePrinter

```sh
kubectl apply -f example.yaml
```

Start the controller

```sh
go run main.go
```
