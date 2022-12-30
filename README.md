# go-genkubeconfig

This tool should help create `~/.kube/config` files for use with EKS clusters in one or many AWS accounts.

## Usage

Get information from a cluster named "dev1" using an AWS profile called "dev":

```
./go-genkubeconfig -clusters dev:dev1
```

You can pass multiple profiles/clusters and redirect STDOUT directly to the `~/.kube/config`:

```
./go-genkubeconfig -clusters dev:dev1 -clusters tst:tst1 -clusters lab:lab1 -clusters stg:stg1 -clusters prd:prd1 > ~/.kube/config
```
