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
        client [--d] [--c=<config>] CLIENT POLICY SCHEDULE SCHEDULETYPE STATUS STREAM PATHNAME
        client --p

Options:
  -h --help
  -c,--c=<config>   JSON configuration file [default: clientConfig.json]
  -p,--p            Print out sample clientConfig.json
  -d,--d            Dry run`

        arguments, _ := docopt.Parse(usage, nil, true, "", false)

        // exit if we are only showing a sample config
        if arguments["--p"].(bool) {
                b := []byte(`{"TLSCommonCA":"./certs/CA.crt","TLSMyCert":"./certs/boxname.crt","TLSMyKey":"./certs/boxname.key","ServerIP": "192.168.0.2","ServerPort":8075,"ClientIP":"192.168.0.3"}`)
                PrintSampleConfig(b)
                return
        }

        // Define the Params for the RPC call
        var args Args
        args.Client = arguments["CLIENT"].(string)
        args.Policy = arguments["POLICY"].(string)
        args.Schedule = arguments["SCHEDULE"].(string)
        args.SchedType = arguments["SCHEDULETYPE"].(string)
        args.Status = arguments["STATUS"].(string)
        args.Stream = arguments["STREAM"].(string)
        args.Path = arguments["PATHNAME"].(string)
        args.DryRun = arguments["--d"].(bool)

        // Read configuration from json file and set defaults
        Settings := ParseConfig(arguments["--c"].(string))

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
                log.Fatal("Connection refused")
        }
        os.Exit(reply.Code)
}
