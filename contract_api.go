package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	//"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
)

// contains context that is required by handlers to interact with
// geth
type quorumContext struct {
	session *SimpleStorageSession // session for deploying/managing contract
	conn    *ethclient.Client     // maintain a rpc client to geth
	key     string                // key for an account we use
	url     string                // url to connect to
}

// create a struct for the handler so that each handler has access to
// connection information required to interact with the chain
type quorumHandler struct {
	*quorumContext
	H func(*quorumContext, http.ResponseWriter, *http.Request) (int, error)
}

// ServeHTTP func that has the ability to
// access *quorumContext's fields - session, key and url
func (h quorumHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// invoke the handler function in the struct
	status, err := h.H(h.quorumContext, w, r)
	if err != nil {
		log.Printf("HTTP status %d: err %q", status, err)
		switch status {
		case http.StatusNotFound:
			http.NotFound(w, r)
		case http.StatusInternalServerError:
			http.Error(w, http.StatusText(status), status)
		default:
			http.Error(w, http.StatusText(status), status)
		}
	}
}

// Deploy a contract to a cluster of quorum nodes
func deploy(ctx *quorumContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// connect to geth; docker container for quorum is
	// port 8545 to port 22001
	conn, err := ethclient.Dial(ctx.url)
	if err != nil {
		fmt.Fprintf(w, "Failed to connect to the Ethereum client: %v", err)
		return 500, err
	}
	auth, err := bind.NewTransactor(strings.NewReader(ctx.key), "")
	if err != nil {
		fmt.Fprintf(w, "Failed to create authorized transactor: %v", err)
		return 500, err
	}

	// Deploy simple storage contract to quorum local node
	address, _, _, err := DeploySimpleStorage(auth, conn)
	if err != nil {
		fmt.Fprintf(w, "Failed to deploy new storage contract: %v", err)
		return 500, err
	}

	//address := common.HexToAddress("0xfd8ff156e8e92fbbd4a9e7873d163ea27775dbe6")

	// Get a handle to the contract
	contract, err := NewSimpleStorage(address, conn)
	if err != nil {
		fmt.Fprintf(w, "Failed to get contract handle: %v", err)
		return 500, err
	}

	session := &SimpleStorageSession{
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
	// set the session, connection in context (bad pattern)
	ctx.conn = conn
	ctx.session = session
	return 200, nil
}

// Display summary of a transaction using its hash
func transactionSummary(ctx *quorumContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if ctx.conn == nil {
		fmt.Fprintf(w, "Cannot get transaction before deploying contract\n")
		return 400, nil
	}

	// parse hash from request URL
	keys := r.URL.Query()["hash"]
	if len(keys) < 1 {
		fmt.Fprintf(w, "Invalid parameter, require 'hash'")
		return 400, nil
	}
	log.Println("Hash supplied :", keys[0])
	hash := common.HexToHash(keys[0])

	basic_context := context.Background()
	tx, pending, err := ctx.conn.TransactionByHash(basic_context, hash)
	//common.HexToHash("0x378674bebd1430d9ce63adc792c573da56e69b8d6c97174c93a43c5991ae0d61"))
	if err != nil {
		fmt.Fprintf(w, "Failed to get transaction details: %v", err)
		return 500, err
	}
	fmt.Fprintf(w, "Transaction pending? %v; details: %v\n",
		pending, tx.String())
	return 200, nil
}

// Get stored value from contract
func get(ctx *quorumContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if ctx.session == nil {
		fmt.Fprintf(w, "Cannot get value before deploying contract\n")
		return 400, nil
	}

	val, err := ctx.session.Get()
	if err != nil {
		fmt.Fprintf(w, "Failed to get stored contract value: %v", err)
		return 500, err
	}
	fmt.Fprintf(w, "Stored value: %v\n", val)
	return 200, nil
}

// Set value in contract
func set(ctx *quorumContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if ctx.session == nil {
		fmt.Fprintf(w, "Cannot set value before deploying contract\n")
		return 400, nil
	}

	// parse value to set from request URL
	keys, _ := r.URL.Query()["data"]
	if len(keys) < 1 {
		fmt.Fprintf(w, "Invalid parameter, require 'data'")
		return 400, nil
	}
	val := new(big.Int)
	_, err := fmt.Sscan(keys[0], val)

	tx, err := ctx.session.Set(val)
	if err != nil {
		fmt.Fprintf(w, "Failed to set/store contract value: %v", err)
		return 500, err
	}
	fmt.Fprintf(w, "Hash of Transaction for Set: 0x%x\n\n", tx.Hash())
	return 200, nil
}

func main() {
	// parse hostname and port
	var hostname string
	var port, listen_port int

	flag.StringVar(&hostname, "host", "localhost",
		"hostname of quorum node")
	flag.IntVar(&port, "port", 22001, "geth rpc port")
	flag.IntVar(&listen_port, "listen_port", 8080,
		"port on which to listen for incoming requests")
	flag.Parse()

	// initialize app context supplied to handlers.
	// this includes state such as the session established to geth,
	// connection, key for the account and the url used to connect to geth
	app_context := &quorumContext{
		session: nil,
		conn:    nil,
		key:     `{"address":"f29f27dacb6c2b616c2552cb0c7a3c7ff5b64d16","crypto":{"cipher":"aes-128-ctr","ciphertext":"3513e917da2e63563811be8f879b4cb127d3da5619f1bd9204bcdfc3725a688e","cipherparams":{"iv":"dcae036447cb85f851354231dc01b252"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"28d7aed161d7f2a1307c0fee4ac10ff16e491a74d7ae29d07a0b38fb75b9d479"},"mac":"7ae92e733a9d290676a685dee20ab36dd533fbf82480f5caf510c83baef123ad"},"id":"2e377c26-92d3-4671-b01f-1dd32c6e2607","version":3}`,
		url:     "http://" + hostname + ":" + strconv.Itoa(port),
	}

	router := mux.NewRouter()
	router.Handle("/deploy",
		quorumHandler{app_context, deploy})
	router.Handle("/get",
		quorumHandler{app_context, get})
	router.Handle("/set",
		quorumHandler{app_context, set})
	router.Handle("/summary",
		quorumHandler{app_context, transactionSummary})

	log.Println("Listening...")
	http.ListenAndServe(":8080", router)
}
