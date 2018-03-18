package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const key = `{"address":"f29f27dacb6c2b616c2552cb0c7a3c7ff5b64d16","crypto":{"cipher":"aes-128-ctr","ciphertext":"3513e917da2e63563811be8f879b4cb127d3da5619f1bd9204bcdfc3725a688e","cipherparams":{"iv":"dcae036447cb85f851354231dc01b252"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"28d7aed161d7f2a1307c0fee4ac10ff16e491a74d7ae29d07a0b38fb75b9d479"},"mac":"7ae92e733a9d290676a685dee20ab36dd533fbf82480f5caf510c83baef123ad"},"id":"2e377c26-92d3-4671-b01f-1dd32c6e2607","version":3}`

var session *SimpleStorageSession

// url to connect to
var url string

// Deploy a contract to a cluster of quorum nodes
func deploy() error {
	// connect to geth; docker container for quorum is
	// port 8545 to port 22001
	conn, err := ethclient.Dial(url)
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

// Display summary of a transaction using its hash
func transactionSummary(hash common.Hash) {
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		return
	}
	ctx := context.Background()
	tx, pending, err := conn.TransactionByHash(ctx, hash)
	//common.HexToHash("0x378674bebd1430d9ce63adc792c573da56e69b8d6c97174c93a43c5991ae0d61"))
	if err != nil {
		log.Fatalf("Failed to get transaction details: %v", err)
		return
	}
	fmt.Printf("Transaction pending %v details: %v\n",
		pending, tx.String())
}

// Get stored value from contract
func get() (*big.Int, error) {
	val, err := session.Get()
	if err != nil {
		log.Fatalf("Failed to get storage: %v", err)
		return nil, err
	}
	fmt.Printf("Stored value: %v\n", val)
	return val, err
}

// Set value in contract
func set(val *big.Int) (common.Hash, error) {
	tx, err := session.Set(val)
	if err != nil {
		log.Fatalf("Set Failed to set storage: %v", err)
		return common.StringToHash(""), err
	}
	fmt.Printf("Hash of Transaction for Set: 0x%x\n\n", tx.Hash())
	return tx.Hash(), err
}

func main() {
	// parse hostname and port
	var hostname string
	var port int

	flag.StringVar(&hostname, "host", "localhost",
		"hostname of quorum node")
	flag.IntVar(&port, "port", 22001, "geth rpc port")
	flag.Parse()
	url = "http://" + hostname + ":" + strconv.Itoa(port)

	deploy()
	time.Sleep(250 * time.Millisecond)
	hash, _ := set(big.NewInt(100))
	time.Sleep(500 * time.Millisecond)
	get()
	transactionSummary(hash)
}
