# go-genkubeconfig

This tool should help create `~/.kube/config` files for use with EKS cluster(s) in one or many AWS accounts.

## Usage

```
./bin/genkubeconfig_darwin -h
Usage of ./bin/genkubeconfig_darwin:
  -cluster value
        AWS Profile and EKS Cluster name joined by a colon, can be passed more than once
        e.g. -cluster dev:dev1 -cluster tst:tst1
  -version
        print current version and exit
```

### Examples

Get information from a single cluster named "dev1" using an AWS profile called "dev":

```
./go-genkubeconfig -cluster dev:dev1
```

Pass multiple profiles/cluster and redirect STDOUT directly to the `~/.kube/config`:

```
./go-genkubeconfig -cluster dev:dev1 -cluster tst:tst1 -cluster lab:lab1 -cluster stg:stg1 -cluster prd:prd1 > ~/.kube/config
```
