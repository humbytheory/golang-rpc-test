package main

import (
	"crypto/tls"
	"log"
	"net/rpc/jsonrpc"
	"os"
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
MIIDMzCCAhugAwIBAgIBAjANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQDEwdNeVJQ
Q0NBMB4XDTE1MDExMTE3MTI0N1oXDTI1MDEwODE3MTI0N1owEjEQMA4GA1UEAxMH
Y2xpZW50MDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANDwTUU1p9Qd
VdKl1aFwJxibArcFMS+E5iUq5lzzRqeJRyJXg9g2BRzm8b74bPKEiQ1zPsadiYzB
PpuiZwytAW7gyj1PqQUO1dWGylVt7BxoYt3xLGPii/fpt0Yn1PtHy75R2mkUl5H3
cEkao4t7fdmvGnumC+X1Dd0Ss9S4G+hHSWNiBbeQMO+gaDsTw7mV7dS6Meh7gTW2
xWpFzWsuYQxfWR/tTgWqlpWXRMt1KE7EM+xelYgLEOUvEoh/LfjEXZ1sCcC5kEN7
AxkJZI/sTpOHA7ah/aG6v8IskL3m4R2qg5pxTP3MOfo1KrLkZdJV0NqyrtskfXUw
viR0+BNHBQECAwEAAaOBkzCBkDAJBgNVHRMEAjAAMB0GA1UdDgQWBBQjL53RWpi0
ePdOKeCjJFXiGHarBDBCBgNVHSMEOzA5gBTqhpFUpA5NXkiT29TPqwtAYhK7UKEW
pBQwEjEQMA4GA1UEAxMHTXlSUENDQYIJAJq7PwIyFW4uMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAsGA1UdDwQEAwIHgDANBgkqhkiG9w0BAQsFAAOCAQEAbjBd1YbKps92
meVcjoro0aNXNNJONCQjaWXWVnJQPG1dETG6z35L7sJk+ZthjXRy1Esc+UCTVCJ6
N8/DRbH2M9AkVsQOTCWGm1VNHuDF6ttUaEXwQy9o5qQjY1FGMZXl59Aci/nJAzFV
WWJJunO2eh1SuI92bZM6lxI58hhM0p7JSWHg8AxsP1ThI+ilBmjxdyI43cV+myzy
EG6A0r2Wx2IOVpuIcM9YRm7946XbwIhxx9+IoGlmf9v4UREIkuZk9/PILUC1fejy
RjfuHhy0dkokoVmuw6CoGG0FwM+CpyHSY54hTP1I2xziA59nQTxMuyRBIU0LqV2+
6EAt9KjQ7Q==
-----END CERTIFICATE-----`
const myPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0PBNRTWn1B1V0qXVoXAnGJsCtwUxL4TmJSrmXPNGp4lHIleD
2DYFHObxvvhs8oSJDXM+xp2JjME+m6JnDK0BbuDKPU+pBQ7V1YbKVW3sHGhi3fEs
Y+KL9+m3RifU+0fLvlHaaRSXkfdwSRqji3t92a8ae6YL5fUN3RKz1Lgb6EdJY2IF
t5Aw76BoOxPDuZXt1Lox6HuBNbbFakXNay5hDF9ZH+1OBaqWlZdEy3UoTsQz7F6V
iAsQ5S8SiH8t+MRdnWwJwLmQQ3sDGQlkj+xOk4cDtqH9obq/wiyQvebhHaqDmnFM
/cw5+jUqsuRl0lXQ2rKu2yR9dTC+JHT4E0cFAQIDAQABAoIBAQCm2J37vIMOoXZd
Rkw4JIUz4uTiHeVPGwNlfsKCS0qKktcZF2WTjF+82rcFVwA5EZkYAoWIuViT6+UB
B0jfGHBiiGM3XpuMDHK5lm+QlLzNWpZIwUQ/ZzN6f0n5Xel317ddfaO58dWvnDYw
SnN6+NxgrrGpN8mcknnFph+wqGywqHrj5tA+WID5gJk6oBttG23NxcO1IW3c7+UV
qi6tLaUOtUVbrtETfwh2+H5mjk1e/ljxZDYrDLxZfoMY9aKz+Vtc49KzEOdrtfsh
Oo2tbqKVdXPMvP/YfUdWu+tvwqYnjIYeaVVFVqtn19AidCj/cOUPiSHIHfMT/xAb
zEKOaDehAoGBAO7cR/bpiVEPU6TDdsvVb6jr5GO+yQ3xz7La+eb9rDaBlLUz5Tdw
uPlMySZcEvLWqc6E/O/TxtIWx87Ahxk1TWEXpq0vcFZSKLJFvZToHXneKqtNc6Se
zTskJtwZz9yhWK3+j/1ibKHzEmPGer7vaw3IpAR2XgHesR+NUW+j7wl9AoGBAN/u
YDubUPeFvOBy/6mC2alPTsdPZ/wNw1B7Amz9mHYRT39WZ5vZPZeAn8JC573fH29Q
y7/eTRKkrF8Hvi9ixqe6tmKl/U4w7QFhAk3/TvF9kPky3rVq1zaB2ZeCASH2N2lL
EpdQuh1/xxkDlAGtcSSlr8LRBOx/VLgZ3xZPLqDVAoGBALFbUvwdj95meP8AO/dB
9fUBosYFZZg7ErOFMMW5WePm95pMfEhcJJzHzRv0hgVWKyOzT3RsVVatn5L/FdE7
6MbNHu+9J7aQrrMgYZJtf2V790bW7aUwXMcrIsePSu5Rx1z6hcPpDyx5JhB70axw
bZcAgfjmQws0ZWQ+NFem69ipAoGAPZ+Ixf546pTYJGAhMRG8OlaD1F9quzdCX3xq
b3neIeejm+Q4QPAofe+8hyYIRf0H1odCert/NDky4jfsQ3gIORItrLoHGiRmpHGA
w9wVamlmot034m7TaMGVEpeJHkJ2fzhUlmV1wjZuoNiWO1vyfeZGlvMUSszDkKI1
/RqvNz0CgYBvmnkyKPHgQIK0idi7nZS8sQGbpZwTG0HTuCRgpk/2AjyC1ZDYvqtY
o33q8KbuiN/aBHZjJgQFrlkeFhUIgNS2RI9DlW24+Wb1isqK8scheYiNT7b/4X5k
MNCsMpIquXWS17wVsv6g4MQOkihfTarUDkJLWKZwlTh7PkYRuzTDng==
-----END RSA PRIVATE KEY-----`

type Args struct {
	Name  string
	Id    int
	State bool
}

type Results struct {
	Code int
}

func main() {
	var MyAddr = "localhost"
	var MyPort = "51000"
	var MyAddrMyPort = MyAddr + ":" + MyPort

	// temp data to send RPC server
	var args Args
	args.Name = `Volumes`
	args.Id = 10
	args.State = false

	// Setup TLS and tslConnect
	config := MustGetTlsConfiguration(myCaCertificate, myCertificate, myPrivateKey)
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
