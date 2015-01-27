package main

import (
	// "bytes"
	"crypto/tls"
	"github.com/docopt/docopt-go"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
var Settings *ConfigSettings

// RPC method struct
type RPCMethods struct{}

// Used by RPC method to run external command
func shellOut(isClientDryRun bool, filer, volume, snapshot string) int {
	if global_DryRun || isClientDryRun {
		log.Printf("RPC Exec Dry Run: Path:%s CMD:%s Filer:%s Volume:%s Snapshot:%s\n", global_ExternalCMDPath, global_ExternalCMD, filer, volume, snapshot)
		return ErrorDryRun
	} else {
		log.Printf("RPC Exec Run: Path:%s CMD:%s Filer:%s Volume:%s Snapshot:%s\n", global_ExternalCMDPath, global_ExternalCMD, filer, volume, snapshot)
		cmd := exec.Command(global_ExternalCMD, filer, volume, snapshot)
		// cmd.Dir = global_ExternalCMDPath

		os.Setenv("testvar", global_ExternalCMDPath)
		// var env []string
		// env = os.Environ()
		// log.Println("List of Environtment variables : \n")
		// for index, value := range env {
		// 	name := strings.Split(value, "=")
		// 	log.Printf("[%d] %s : %v\n", index, name[0], name[1])
		// }

		log.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("==> Error: %s\n", err.Error())
			return ErrorExternal
		}
		if len(output) > 0 {
			log.Printf("==> Output: %s\n", string(output))
		}
	}
	return ErrorNone
}

// RPC method that will be called by client
func (t *RPCMethods) DoSomething(args *Args, reply *Results) error {
	if global_DebugLevel == "1" {
		log.Printf("RPC: serverDryRun:%v args: %v  \n", global_DryRun, args)
	}

	filer, vol := "", ""
	reply.Code = ErrorInvalidInput

	for _, validClient := range Settings.Filers {
		if validClient.Name == args.Client && validClient.Enabled == 1 {
			filer = validClient.Name
		}
	}
	if filer == "" {
		log.Printf("RPC: Rejected Invalid Filer: %s\n", args.Client)
		return nil
	}
	if args.Status != "0" {
		log.Printf("RPC: Rejected Invalid Status: %s\n", args.Status)
		return nil
	}
	if args.SchedType != "FULL" {
		log.Printf("RPC: Rejected Invalid SchedType: %s\n", args.SchedType)
		return nil
	}

	p := strings.Split(args.Path, "/")
	if len(p) >= 5 {
		if p[len(p)-1] != Settings.SnapshotName {
			log.Printf("RPC: Rejected Invalid Snapshot Name: %s\n", p[len(p)-1])
			return nil
		}
		if p[1] != "vol" || p[len(p)-2] != ".snapshot" {
			log.Printf("RPC: Rejected Invalid Snapshot Path Name: %s\n", p[len(p)-1])
			return nil
		}
		if isValidInput(p[2]) {
			vol = p[2]
		} else {
			log.Printf("RPC: Rejected Invalid Volume: %s\n", p[2])
			return nil
		}
	} else {
		log.Printf("RPC: Rejected Invalid Path: %s\n", args.Path)
		return nil
	}
	reply.Code = shellOut(args.DryRun, filer, vol, Settings.SnapshotName)
	return nil

}

// Used by RPM method to report if input data is valid
func isValidInput(input string) bool {
	if len(input) < 1 {
		return false
	}
	for _, c := range input {
		if c > unicode.MaxASCII || !unicode.IsPrint(c) {
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
		b := []byte(`{"tlscommonca":"/certs/CA.crt","tlscert":"/certs/boxname.crt","tlskey":"/certs/boxname.key","serverip": "192.168.0.2","serverport":8075,"clientip":"192.168.0.3","externalcmdpath":"/tmp","externalcmd":"command"}`)
		PrintSampleConfig(b)
		return
	}

	// Read configuration from json file and set defaults
	Settings = ParseConfig(arguments["--c"].(string))

	global_ExternalCMDPath = Settings.ExternalCMDPath
	global_ExternalCMD = Settings.ExternalCMD
	global_DryRun = arguments["--t"].(bool)
	global_DebugLevel = arguments["--d"].(string)

	configIpList := make(map[string]bool)
	configIpList[Settings.ClientIP] = true

	// Setup TLS and start listener
	config := GetTLSConfig(Settings.TLSCommonCA, Settings.TLSCert, Settings.TLSKey)

	listener, err := tls.Listen("tcp", Settings.ServerIPPort, config)
	defer listener.Close()
	if err != nil {
		log.Fatalf("TLS Error: Failed to start listening: %v\n", err)
	}

	// Setup RPC server
	server := rpc.NewServer()
	cal := new(RPCMethods)
	server.Register(cal)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

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
