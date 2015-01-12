package main

import (
        "crypto/tls"
        "log"
        "net/rpc/jsonrpc"
        "os"
        "strconv"
)

type Args struct {
        Name  string
        Id    int
        State bool
}

type Results struct {
        Code int
}

func main() {
        Settings := ParseConfig("clientConfig.json")
        log.Printf("Settings:")
        log.Printf("TLSCommonCA: %s\n", Settings.TLSCommonCA)
        log.Printf("TLSMyCert: %s\n", Settings.TLSMyCert)
        log.Printf("TLSMyKey: %s\n", Settings.TLSMyKey)
        log.Printf("ServerIP: %s\n", Settings.ServerIP)
        log.Printf("ServerPort: %s\n", Settings.ServerPort)
        log.Printf("ExternalCMDPath: %s\n", Settings.ExternalCMDPath)
        log.Printf("ExternalCMD: %s\n", Settings.ExternalCMD)
        log.Printf("ClientIP: %s\n", Settings.ClientIP)

        var MyAddr = Settings.ServerIP
        var MyPort = strconv.Itoa(Settings.ServerPort)
        var MyAddrMyPort = MyAddr + ":" + MyPort

        // temp data to send RPC server
        var args Args
        args.Name = `Volumes`
        args.Id = 10
        args.State = false

        // Setup TLS and tslConnect
        config := MustGetTlsConfiguration(Settings.TLSCommonCA, Settings.TLSMyCert, Settings.TLSMyKey)
        tslConn, err := tls.Dial("tcp", MyAddrMyPort, config)
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
