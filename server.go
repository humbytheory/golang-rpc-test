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
var global_DebugLevel string

// RPC method struct
type RPCMethods struct{}

// Used by RPC method to run external command
func shellOut(what string) int {
	if global_DebugLevel == "2" {
		log.Printf("RPC: [%v] c:%s p:%s", what, global_ExternalCMDPath, global_ExternalCMD)
	}
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
	if global_DebugLevel == "2" {
		log.Printf("RPC: args: %v  serverDryRun: %v\n", args, global_DryRun)
	}
	if global_DryRun || args.DryRun {
		reply.Code = ErrorDryRun
		return nil
	}
	if isValidInput(args.Client) {
		reply.Code = shellOut(args.Client)
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
        server [--t] [--d=<level>] [--c=<config>]
        server --p

Options:
  -h --help
  -c,--c=<config>   JSON configuration file [default: serverConfig.json]
  -p,--p            Print out sample serverConfig.json
  -d,--d=<level>    Debug level 1-3. [default: 0]
  -t,--t            Test run`

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
	global_DryRun = arguments["--t"].(bool)
	global_DebugLevel = arguments["--d"].(string)

	configIpList := make(map[string]bool)
	configIpList[Settings.ClientIP] = true

	// Setup TLS and start listener
	config := MustGetTlsConfiguration(Settings.TLSCommonCA, Settings.TLSMyCert, Settings.TLSMyKey)
	listener, err := tls.Listen("tcp", Settings.ServerIPPort, config)
	defer listener.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Setup RPC server
	server := rpc.NewServer()
	cal := new(RPCMethods)
	server.Register(cal)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	//https://groups.google.com/d/topic/golang-nuts/l09ZJQa5Cnk
	//https://gist.github.com/2232102
	//https://gist.github.com/2233075
	//https://gist.github.com/1866829

	// Handle connection from client
	for {
		conn, err := listener.Accept()
		defer conn.Close()
		if err != nil {
			log.Println(err)
		}

		tlscon, ok := conn.(*tls.Conn)
		if ok {
			if global_DebugLevel == "1" {
				log.Println("Server: conn: type assert to TLS succeedded")
			}
			err := tlscon.Handshake()
			if err != nil {
				conn.Close()
				log.Println("Server: handshake failed: %s", err)
			}
			state := tlscon.ConnectionState()
			if global_DebugLevel == "1" {
				log.Print("Server: conn: Handshake completed")
				log.Printf("Client: Requested server name: %v\n", state.ServerName)
				log.Printf("Client: Cert CommonName: %v\n", state.PeerCertificates[0].Subject.CommonName)
				log.Printf("Client: Cert DNS Names %v\n", state.PeerCertificates[0].DNSNames)
				log.Printf("Client: Cert Email Addresses %v\n", state.PeerCertificates[0].EmailAddresses)
				log.Printf("Client: Cert IP Addresses %v\n", state.PeerCertificates[0].IPAddresses)
			}

			if global_DebugLevel == "3" {
				// Use this to debug RPC method called and params provided
				buf := make([]byte, 512)
				for {
					log.Print("Server: conn: waiting")
					n, err := conn.Read(buf)
					if err != nil {
						if err != nil {
							log.Printf("Server: conn: read: %s", err)
						}
						break
					}
					log.Printf("Server: conn: echo %q\n", string(buf[:n]))
					n, err = conn.Write(buf[:n])
					log.Printf("Server: conn: wrote %d bytes", n)
					if err != nil {
						log.Printf("Server: write: %s", err)
						break
					}
				}
				log.Printf("Server: Debug 3 completed. Quitting.")
				return
			}

			clientIpPort := conn.RemoteAddr().String()
			clientIp, _, err := net.SplitHostPort(clientIpPort)
			if err != nil {
				conn.Close()
				log.Printf("Server: Error getting client IP: %v\n", err)
			}
			if clientIp != state.PeerCertificates[0].IPAddresses[0].String() {
				conn.Close()
				log.Printf("Server: Rejected connection from: %v\n", clientIpPort)
			}
			if configIpList[clientIp] {
				if global_DebugLevel == "1" {
					log.Printf("Server: Accepted connection from: %v\n", clientIpPort)
				}
				go server.ServeCodec(jsonrpc.NewServerCodec(conn))
			} else {
				conn.Close()
				log.Printf("Server: Rejected connection from: %v\n", clientIpPort)
			}
		} else {
			//There was a problem with the TLS connection dropping client
			conn.Close()
			log.Println("Server: TLS conn error: %s", err)
		}
	}
}
