package main

import (
        "bytes"
        "crypto/tls"
        "log"
        "net"
        "net/rpc"
        "net/rpc/jsonrpc"
        "os/exec"
        "regexp"
        "strconv"
        "unicode"
)

type Args struct {
        Name  string
        Id    int
        State bool
}

type Results struct {
        Code int
}

type RPCMethods struct {
        // just an empty struct
}

const ErrorNone = 0
const ErrorInvalidInput = 1
const ErrorExternal = 2

func shellOut(what string) int {
        log.Printf("[%v]", what)
        cmd := exec.Command("echo", "-n", "wakka")
        // cmd.Dir = "/tmp"
        var out bytes.Buffer
        cmd.Stdout = &out
        if err := cmd.Run(); err != nil {
                return ErrorExternal
        }
        return ErrorNone
}

func (t *RPCMethods) DoSomething(args *Args, reply *Results) error {
        if isValidInput(args.Name) {
                reply.Code = shellOut(args.Name)
        } else {
                reply.Code = ErrorInvalidInput
        }
        return nil
}

func isValidInput(input string) bool {
        for _, character := range input {
                if character > unicode.MaxASCII || !unicode.IsPrint(character) {
                        return false
                }
        }
        re := regexp.MustCompile("^[a-zA-Z0-9_-]*$")
        if re.MatchString(input) {
                return true
        } else {
                return false
        }
}

func main() {
        Settings := ParseConfig("serverConfig.json")
        log.Printf("Settings:")
        log.Printf("TLSCommonCA: %s\n", Settings.TLSCommonCA)
        log.Printf("TLSMyCert: %s\n", Settings.TLSMyCert)
        log.Printf("TLSMyKey: %s\n", Settings.TLSMyKey)
        log.Printf("ServerIP: %s\n", Settings.ServerIP)
        log.Printf("ServerPort: %s\n", Settings.ServerPort)
        log.Printf("ExternalCMDPath: %s\n", Settings.ExternalCMDPath)
        log.Printf("ExternalCMD: %s\n", Settings.ExternalCMD)
        log.Printf("ClientIP: %s\n", Settings.ClientIP)

        configIpList := make(map[string]bool)
        configIpList[Settings.ClientIP] = true

        var MyAddr = Settings.ServerIP
        var MyPort = strconv.Itoa(Settings.ServerPort)
        var MyAddrMyPort = MyAddr + ":" + MyPort

        // Setup TLS and start listener
        config := MustGetTlsConfiguration(Settings.TLSCommonCA, Settings.TLSMyCert, Settings.TLSMyKey)
        listener, err := tls.Listen("tcp", MyAddrMyPort, config)
        defer listener.Close()
        if err != nil {
                log.Fatal(err)
        }

        // Setup RPC
        server := rpc.NewServer()
        cal := new(RPCMethods)
        server.Register(cal)
        server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

        // Handle connection from client
        for {
                conn, err := listener.Accept()
                if err != nil {
                        log.Fatal(err)
                }
                clientIpPort := conn.RemoteAddr().String()
                clientIp, _, err := net.SplitHostPort(clientIpPort)
                if err != nil {
                        log.Fatal(err)
                }
                if configIpList[clientIp] {
                        log.Printf("Accepted connection from: %v\n", clientIpPort)
                        go server.ServeCodec(jsonrpc.NewServerCodec(conn))
                } else {
                        log.Printf("Rejected connection from: %v\n", clientIpPort)
                }

        }
}
