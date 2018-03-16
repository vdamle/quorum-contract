package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	//	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const key = `{"address":"f29f27dacb6c2b616c2552cb0c7a3c7ff5b64d16","crypto":{"cipher":"aes-128-ctr","ciphertext":"3513e917da2e63563811be8f879b4cb127d3da5619f1bd9204bcdfc3725a688e","cipherparams":{"iv":"dcae036447cb85f851354231dc01b252"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"28d7aed161d7f2a1307c0fee4ac10ff16e491a74d7ae29d07a0b38fb75b9d479"},"mac":"7ae92e733a9d290676a685dee20ab36dd533fbf82480f5caf510c83baef123ad"},"id":"2e377c26-92d3-4671-b01f-1dd32c6e2607","version":3}`

var session *SimpleStorageSession

func deploy() error {
	// connect to geth
	conn, err := ethclient.Dial("geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		return err
	}
	auth, err := bind.NewTransactor(strings.NewReader(key), "")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
		return err
	}

	// Deploy simple storage contract to quorum local node
	address, _, _, err := DeploySimpleStorage(auth, conn)
	if err != nil {
		log.Fatalf("Failed to deploy new storage contract: %v", err)
		return err
	}

	//address := common.HexToAddress("0xfd8ff156e8e92fbbd4a9e7873d163ea27775dbe6")

	// Get a handle to the contract
	contract, err := NewSimpleStorage(address, conn)
	if err != nil {
		log.Fatalf("Failed to get contract handle: %v", err)
		return err
	}

	session = &SimpleStorageSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From:     auth.From,
			Signer:   auth.Signer,
			GasLimit: uint64(10000000),
		},
	}
	return err
}

func get() (*big.Int, error) {
	val, err := session.Get()
	if err != nil {
		log.Fatalf("Failed to get storage: %v", err)
		return nil, err
	}
	fmt.Printf("Stored value: %v\n", val)
	return val, err
}

func set(val *big.Int) error {
	tx, err := session.Set(val)
	if err != nil {
		log.Fatalf("Set Failed to set storage: %v", err)
		return err
	}
	fmt.Printf("Hash of Transaction for Set: 0x%x\n\n", tx.Hash())
	return err
}

func main() {
	deploy()
	time.Sleep(250 * time.Millisecond)
	set(big.NewInt(100))
	time.Sleep(500 * time.Millisecond)
	get()
}
