# Quorum Contract
This repo contains code that implements APIs to interact with a quorum
cluster to deploy and interact with a smart contract

## Steps

* Build the go program wherever the source file is located
```
cd <repo root>
go build .
```

## Prerequisites

* Quorum nodes have been instantiated

* `go-ethereum` is installed and binaries are accessible in runtime path

* `abigen` is installed in the runtime path so that go binding code can
  be generated from the smart contract. Example command:

```
cd ~/go/src/github.com/vdamle/quorum-contract
abigen --sol storage.sol --pkg main --out storage.go
```
* Key for an account on a node that can be found in /qdata/ethereum (within the container)
  is included in the go program that interacts with the blockchain
