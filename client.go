package main

import (
	"crypto/tls"
	"github.com/docopt/docopt-go"
	"log"
	"net/rpc/jsonrpc"
	"os"
)

func main() {
	usage := `
Usage:
	client [--d] [--c=<config>] ONE TWO THREE FOUR FIVE SIX
	client --p

Options:
  -h --help
  -c,--c=<config>   JSON configuration file [default: clientConfig.json]
  -p,--p            Print out sample clientConfig.json
  -d,--d            Dry run`

	arguments, _ := docopt.Parse(usage, nil, true, "", false)

	// Read configuration from json file and set defaults
	Settings := ParseConfig(arguments["--c"].(string))
	log.Fatal(arguments)

	// temp data to send RPC server
	var args Args

	// Setup TLS and tslConnect
	config := MustGetTlsConfiguration(Settings.TLSCommonCA, Settings.TLSMyCert, Settings.TLSMyKey)
	tslConn, err := tls.Dial("tcp", Settings.ServerIPPort, config)
	defer tslConn.Close()
	if err != nil {
		log.Fatal("Error dialing:", err)
	}
	err = tslConn.Handshake()
	if err != nil {
		log.Fatal("Failed handshake:%v\n", err)
	}

	// Do a synchronous call
	var reply Results
	client := jsonrpc.NewClient(tslConn)
	defer client.Close()
	err = client.Call("RPCMethods.DoSomething", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	// log.Printf("Sent: %v   Recieved: %v\n", args, reply)
	// log.Printf("Recieved: %v\n", reply.Code)
	os.Exit(reply.Code)
}
