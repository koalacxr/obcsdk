package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"obcsdk/chaincode"
	"obcsdk/peernetwork"
)

var peerNetworkSetup peernetwork.PeerNetwork
var AVal, BVal, curAVal, curBVal, invokeValue int64
var argA = []string{"a"}
var argB = []string{"counter"}
var counter int64
var wg sync.WaitGroup

const (
	TRX_COUNT = 20000
	CLIENTS   = 2
)

func initNetwork() {
	fmt.Println("========= Init Network =========")
	peernetwork.GetNC_Local()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	fmt.Println("========= Register Users =========")
	chaincode.RegisterUsers()
}

func invokeChaincode(peer string) {
	counter++
	arg1Construct := []string{"mycc", "invoke", peer}
	arg2Construct := []string{"a" + strconv.FormatInt(counter, 10), DATA, "counter"}

	_, _ = chaincode.InvokeOnPeer(arg1Construct, arg2Construct)
}

func Init() {
	//initialize
	done := make(chan bool, 1)
	counter = 0
	wg.Add(CLIENTS)
	// Setup the network based on the NetworkCredentials.json provided
	initNetwork()

	//Deploy chaincode
	deployChaincode(done)
}

func InvokeLoop() {
	curTime := time.Now()
	go func() {
		for i := 1; i <= TRX_COUNT/CLIENTS; i++ {
			if counter%1000 == 0 {
				elapsed := time.Since(curTime)
				fmt.Println("=========>>>>>> Iteration#", counter, " Time: ", elapsed, "CLIENT-1")
				curTime = time.Now()
			}
			invokeChaincode("PEER0")
		}
		wg.Done()
	}()
	go func() {
		for i := 1; i <= TRX_COUNT/CLIENTS; i++ {
			if counter%1000 == 0 {
				elapsed := time.Since(curTime)
				fmt.Println("=========>>>>>> Iteration#", counter, " Time: ", elapsed, "CLIENT-2")
				curTime = time.Now()
			}
			invokeChaincode("PEER1")
		}
		wg.Done()
	}()
}

//Cleanup methods to display useful information
func tearDown() {
	fmt.Println("....... State transfer is happening, Lets take a nap for 2 mins ......")
	sleep(120)
	val1, val2 := queryChaincode(counter)
	fmt.Printf("\n========= After Query values a%d = %s,  counter = %s\n", counter, val1, val2)

	newVal, err := strconv.ParseInt(val2, 10, 64)

	if err != nil {
		fmt.Println("Failed to convert ", val2, " to int64\n Error: ", err)
	}

	//TODO: Block size again depends on the Block configuration in pbft config file
	//Test passes when 2 * block height match with total transactions, else fails
	if newVal == counter {
		fmt.Println("\n######### Inserted ", counter, " records #########\n")
		fmt.Println("######### TEST PASSED #########")
	} else {
		fmt.Println("######### TEST FAILED #########")
	}

}

//Execution starts here ...
func main() {
	//TODO:Add support similar to GNU getopts, http://goo.gl/Cp6cIg
	if len(os.Args) < 1 {
		fmt.Println("Usage: go run LedgerStressTwoCliTwoPeer.go Utils.go")
		return
	}
	//TODO: Have a regular expression to check if the give argument is correct format
	/*if !strings.Contains(os.Args[1], "http://") {
		fmt.Println("Error: Argument submitted is not right format ex: http://127.0.0.1:5000 ")
		return;
	}*/
	//Get the URL
	//url := os.Args[1]

	// time to messure overall execution of the testcase
	defer TimeTracker(time.Now(), "Total execution time for LedgerStressTwoCliTwoPeer.go ")

	Init()
	fmt.Println("========= Transacations execution stated  =========")
	InvokeLoop()
	wg.Wait()
	fmt.Println("========= Transacations execution ended  =========")
	tearDown()
}
