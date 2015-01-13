package main

import (
	"bytes"
	"crypto/tls"
	"github.com/docopt/docopt-go"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"regexp"
	"unicode"
)

// RPC status codes to sent to client
const ErrorNone = 0
const ErrorInvalidInput = 1
const ErrorExternal = 2
const ErrorDryRun = 4

// Global vars to allow RPC methods to access config file params
var global_ExternalCMDPath string
var global_ExternalCMD string
var global_DryRun bool

// RPC method struct
type RPCMethods struct{}

// Used by RPC method to run external command
func shellOut(what string) int {
	log.Printf("[%v] c:%s p:%s", what, global_ExternalCMDPath, global_ExternalCMD)
	cmd := exec.Command("echo", "-n", "wakka")
	cmd.Dir = global_ExternalCMDPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ErrorExternal
	}
	// log.Println(cmd.Stdout)
	return ErrorNone
}

// RPC method that will be called by client
func (t *RPCMethods) DoSomething(args *Args, reply *Results) error {
	log.Println(args)
	if global_DryRun || args.DryRun {
		reply.Code = ErrorDryRun
		return nil
	}
	if isValidInput(args.ClientName) {
		reply.Code = shellOut(args.ClientName)
	} else {
		reply.Code = ErrorInvalidInput
	}
	return nil
}

// Used by RPM method to report if input data is valid
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
	usage := `
Usage:
	server [--d] [--c=<config>]
	server --p

Options:
  -h --help
  -c,--c=<config>   JSON configuration file [default: serverConfig.json]
  -p,--p            Print out sample serverConfig.json
  -d,--d            Dry run`

	arguments, _ := docopt.Parse(usage, nil, true, "", false)

	// Exit if we are only showing a sameple config
	if arguments["--p"].(bool) {
		b := []byte(`{"TLSCommonCA":"./certs/CA.crt","TLSMyCert":"./certs/boxname.crt","TLSMyKey":"./certs/boxname.key","ServerIP": "192.168.0.2","ServerPort":8075,"ClientIP":"192.168.0.3","ExternalCMDPath":"/tmp","ExternalCMD":"ls"}`)
		PrintSampleConfig(b)
		return
	}

	// Read configuration from json file and set defaults
	Settings := ParseConfig(arguments["--c"].(string))

	global_ExternalCMDPath = Settings.ExternalCMDPath
	global_ExternalCMD = Settings.ExternalCMD
	global_DryRun = arguments["--d"].(bool)

	configIpList := make(map[string]bool)
	configIpList[Settings.ClientIP] = true

	// Setup TLS and start listener
	config := MustGetTlsConfiguration(Settings.TLSCommonCA, Settings.TLSMyCert, Settings.TLSMyKey)
	listener, err := tls.Listen("tcp", Settings.ServerIPPort, config)
	defer listener.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Setup RPC
	server := rpc.NewServer()
	cal := new(RPCMethods)
	server.Register(cal)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	// RPC Handle connection from client
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		clientIpPort := conn.RemoteAddr().String()
		clientIp, _, err := net.SplitHostPort(clientIpPort)
		if err != nil {
			conn.Close()
			log.Fatal(err)
		}
		if configIpList[clientIp] {
			log.Printf("Accepted connection from: %v\n", clientIpPort)
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		} else {
			log.Printf("Rejected connection from: %v\n", clientIpPort)
			conn.Close()
		}
	}
}
