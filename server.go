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
	"unicode"
)

const myCaCertificate = `-----BEGIN CERTIFICATE-----
MIIDKDCCAhCgAwIBAgIJAJq7PwIyFW4uMA0GCSqGSIb3DQEBCwUAMBIxEDAOBgNV
BAMTB015UlBDQ0EwHhcNMTUwMTExMTcxMjA2WhcNMjUwMTA4MTcxMjA2WjASMRAw
DgYDVQQDEwdNeVJQQ0NBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
oHepokfQ+k7XqoWsWtxQqf0nm//22XZQ3U5rXofgVLUQ83B5eJ9S2J5/RBP0kJjK
0tFyyBed21cZ3FT2Hp5cTudcm4NWyu4eNauMi0Ab1Ju1eUv5K5oRumEA7Mb+WjSY
MAeiZeorHwGHvZEw+He544f2qyCNZg3zun2bNMgTMUmiVY3DbbEo1WnQrk9Dtsr2
daGB5/0P7l3qfhl7lbBx89khf2MDKbMGkAMHR5fyJWuoCZVKdzRp8fENwOorv2ZM
LawkrPX1oRAGGfA5pISxl1W+18INHrF0xJWGiocWkgMwjYqUTMWwLqYa4U4fgFdo
IOQrt8WFzlxW1XwphCeqvwIDAQABo4GAMH4wHQYDVR0OBBYEFOqGkVSkDk1eSJPb
1M+rC0BiErtQMEIGA1UdIwQ7MDmAFOqGkVSkDk1eSJPb1M+rC0BiErtQoRakFDAS
MRAwDgYDVQQDEwdNeVJQQ0NBggkAmrs/AjIVbi4wDAYDVR0TBAUwAwEB/zALBgNV
HQ8EBAMCAQYwDQYJKoZIhvcNAQELBQADggEBABW7n53qZ0+2GDpBJxjoMxKBVCF4
Z8rYk7ALAlELBbKnv3jZiF4bdlfiLVNMmW2fCcZPBEtydkoYLmaA30QbVL8S0os4
I53LRlnrBGtUrsF+EPvDLu4SG9D3i2H818Zh2cdefkC5ZQkkMcGdrDI34qv5tk2+
SqA/mJ5KgJhPQ9oXKz0av7umpOcCk4YNuLQzvipzfXLIeCx4TYy/qMfYnj6Jd1MY
ZkH+pxXExhQFr/ZQnSxQ5qNww618h6HMS4z+LKrOD+iviLO1VrSrAbuO/KGRZzWH
7rTdethVWkwH9Xu6ClcUbO3isscA1JGeyEpFpMH4WAFy7FIk+4agNuv5HHI=
-----END CERTIFICATE-----`
const myCertificate = `-----BEGIN CERTIFICATE-----
MIIDNTCCAh2gAwIBAgIBATANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQDEwdNeVJQ
Q0NBMB4XDTE1MDExMTE3MTIzMVoXDTI1MDEwODE3MTIzMVowFDESMBAGA1UEAxMJ
bG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzZtoPqkI
rWq7/FusmQZQe/UUrl5HeTUODxRzGnWvrdwR5nMVGQ9Bga5cBP/3o+SFGtTpCil0
DI02pDMwol9a+GervdV7PWy28+WnH/VCr/L/an65blhKqzcAEbOYIKpb0sZbnmQZ
44HS6fTBZNkJQRKalfavG0a27ZbvrDRDQKvr+ydQh1jjm2F3EIZsa+py8RU29oR9
JarUeLJLJBLujCzIEHZOm0YN76GQUjK/63/edfnvNXBJyyWpWHAEETPINDyPw51j
Maso9K4cwERty49U0SiB9hIlK54LongW5Qc5BEC2VoF7hVboksrovHcES1Yvhnm8
Q4psaH7i14zx1wIDAQABo4GTMIGQMAkGA1UdEwQCMAAwHQYDVR0OBBYEFLJICx4p
R85toHcWiVJv3pQMyO4LMEIGA1UdIwQ7MDmAFOqGkVSkDk1eSJPb1M+rC0BiErtQ
oRakFDASMRAwDgYDVQQDEwdNeVJQQ0NBggkAmrs/AjIVbi4wEwYDVR0lBAwwCgYI
KwYBBQUHAwEwCwYDVR0PBAQDAgWgMA0GCSqGSIb3DQEBCwUAA4IBAQB4OXW9tnEG
wojujA+RAxh7qs7VJeaaAQPFwyNdJl0h5k/Grdenxo7fPYeNgDglo8EYLH7YmIhv
Nx4ybQj6uNmo7L1hEy4NFw8kduoMcvci2U/y2z32ZFaDwdJ4vhLoTAv3x+lXlnJH
H9jOFnENo6ssZGq0g31WPaGizF6Uoc/fi6dl7XBAVM7A9x6rtEuumT96QXpgcOGO
h9KPplCyzqG83q9vgr8Z0uEpsWvElCCVrGp6k/91xQbhuBLcZuSeE0vxoXmiQV/3
FKk+W81dwhIg1o9tYg9TTVHWGYSuKjznKpaOVm5ng/PtpDAwDWGcVZunfnSUJ0EP
ZAdzdDQXj8jH
-----END CERTIFICATE-----`
const myPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAzZtoPqkIrWq7/FusmQZQe/UUrl5HeTUODxRzGnWvrdwR5nMV
GQ9Bga5cBP/3o+SFGtTpCil0DI02pDMwol9a+GervdV7PWy28+WnH/VCr/L/an65
blhKqzcAEbOYIKpb0sZbnmQZ44HS6fTBZNkJQRKalfavG0a27ZbvrDRDQKvr+ydQ
h1jjm2F3EIZsa+py8RU29oR9JarUeLJLJBLujCzIEHZOm0YN76GQUjK/63/edfnv
NXBJyyWpWHAEETPINDyPw51jMaso9K4cwERty49U0SiB9hIlK54LongW5Qc5BEC2
VoF7hVboksrovHcES1Yvhnm8Q4psaH7i14zx1wIDAQABAoIBAHJVpVaS8PxeikL/
R6+gz0jfNKzySJSiaDsCiC+Cmjr4Ugvwmx7gWPEgYJN3M+KzxUDyfNTl0F7aeDQ/
MyBYHmJcZCigenPh7KscXh9rZ7YoTtiNt9ggyQUFBMjTMhmYIo/HNlOSHsNhAkSP
kqvd9UN2cPhLwXxNipP8hzMfrPZcpDTw6qxIJ4uq76Hd1Q4s+6Tz6kGpTD8fGAY+
tYSdqzMoZyzfen8ji+zMw/3LhaUuVMHlW9X0+KCc6kC+Zdcr75yNY0ix0xRzMjvI
Iu9P0w4zgUVyjXSgM2iG/TnxSaidawMhrunduSk4q83arQQcXjiUEuO9ybd57qp8
OIjJlSECgYEA+hgQ2XrHwFBKOsS/+eenWD6SpTqvytVZsjEf7O2QiA91dMzbmCum
PIxwfyi9ndQbDZP7nPH5N9Pdn57y+kg59Z/Kcfn4AnDLUtWZsf3ukXsR9l97c+lQ
uZQoYG0eE4g5xxB+PduJ9c8vx+u2auzYGvzFXf5yT7jTaMDz1el9ZaMCgYEA0nZl
rP0X3leiEnsNvUHL2BBGnSiSpy8cHoFLMpwWyWLQCPbZrzl+Tb8RJGtLWwNGoAwK
iCHMVEQP6AYFxcbwujcuisPLOjcYJ2NBuEj9QFQnDk1YvaOvk+9jpU0iBE3x+Zrh
j4YW1IiX4JJavoA8BLlLI7Dg62xQwD6iJ2Qc/j0CgYAot65WmiTXbLsJImtXFp4q
QdXCTPG+BkpaNqFKA8uaO1oWMBw4hDLGfN779PgaMCRPa551iPfYXQgiKtDIauX0
1ZUyRU5Zp1+TFu+1CPDEgtMD17vTvVLFRBfmyx0wdOdjP44uKAYoHRlcZUYH1pPA
oRLJINofnKnezjtkwmUGHQKBgCWBYectBzbhSQmgEje65PexFtRk6ZWPiKRLCDqR
pGHpEQe37d0TEtYKCaUC1d/3OnvFCY9u7nnJ00fW4up25Gla8hlagPnz3YMPZiPQ
Jglzta4PzJOm+uATFh/cGgbIWSnRFwc7rw/a863ahv9R3OA+oQxQNhTeLZnEz6LT
bXNFAoGBALxdToA5beCqsb9ciAIzni7VlBFPG2jj7Xyrkwoju58T3KdZef4JxA2Z
nuSqvBs4d4911oNgp2C8XeeY8QOrgsKxZ4uTZMYGFOJMB2EJ82+ehQF1k8n1+rbY
Pln05H+QG8gFWMMevRAlACnpHRTknLXnItBjEcuh1oiCD9RwgZIz
-----END RSA PRIVATE KEY-----`

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
	// log.Printf("[%v]", what)
	cmd := exec.Command("echo", "-n", what)
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
	configIpList := make(map[string]bool)
	configIpList["127.0.0.1"] = true
	configIpList["192.168.0.100"] = true

	var MyAddr = "localhost"
	var MyPort = "51000"
	var MyAddrMyPort = MyAddr + ":" + MyPort

	// Setup TLS and start listener
	config := MustGetTlsConfiguration(myCaCertificate, myCertificate, myPrivateKey)
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
