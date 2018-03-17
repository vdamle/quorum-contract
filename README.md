# Quorum Contract
This repo contains code that implements APIs to interact with a quorum
cluster to deploy and interact with a smart contract

## Steps

* Build the go bindings code for the contract (one time only)
```
cd ~/go/src/github.com/vdamle/quorum-contract
abigen --sol storage.sol --pkg main --out storage.go
```

* Build the go program wherever the source file is located
```
cd <repo root>
go build .
```

* Copy the go executable to the path which is visible in docker (`/qdata/ethereum`).
  This will be `build/docker/tests/tmp/qdata_1/ethereum` if you want to interact with
  node1

* Launch a bash shell for docker and run the go executable
```
docker exec -it <id> bash
cd /qdata/ethereum
./quorum-contract
```

## Prerequisites

* Quorum nodes have been instantiated

* `go-ethereum` is installed and binaries are accessible in runtime path

* `abigen` is installed in the runtime path so that go binding code can
  be generated from the smart contract. Example command:

* Key for an account on a node that can be found in /qdata/ethereum (within the container)
  is included in the go program that interacts with the blockchain
