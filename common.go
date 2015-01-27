package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// Used by the RPC methods
type Args struct {
	Client    string
	Policy    string
	Schedule  string
	SchedType string
	Status    string
	Stream    string
	Path      string
	DryRun    bool
}

// Use by RPC server and client for RPC status codes
type Results struct {
	Code int
}

// Struct for server and client configuration
type Filer struct {
	Name    string `json:"name"`
	Enabled int    `json:"enabled"`
}
type ConfigSettings struct {
	SnapshotName    string   `json:"snapshotname`
	TLSCommonCA     string   `json:"tlscommonca"`
	TLSCert         string   `json:"tlscert"`
	TLSKey          string   `json:"tlskey"`
	ServerIPPort    string   `json:"serveripport"`
	ServerIP        string   `json:"serverip"`
	ServerPort      int      `json:"serverport"`
	ExternalCMDPath string   `json:"externalcmdpath"`
	ExternalCMD     string   `json:"externalcmd"`
	ClientIP        string   `json:"clientip"`
	Filers          []*Filer `json:"filers"`
}

func PrintSampleConfig(b []byte) {
	var dat map[string]interface{}
	if err := json.Unmarshal(b, &dat); err != nil {
		panic(err)
	}
	response, _ := json.MarshalIndent(dat, "", "    ")
	fmt.Println(string(response))
}

// Parse config file into struct
func ParseConfig(filename string) (c *ConfigSettings) {
	configFile, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error opening configuration file: \"", filename, "\"   ", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&c); err != nil {
		log.Fatal("Error parsing configuration file: \"", filename, "\"   ", err.Error())
	}
	c.ServerIPPort = c.ServerIP + ":" + strconv.Itoa(c.ServerPort)

	return c
}

func GetTLSConfig(caFile, certificateFile, privateKeyFile string) *tls.Config {
	config := &tls.Config{}

	mycert, err := tls.LoadX509KeyPair(certificateFile, privateKeyFile)
	if err != nil {
		log.Fatalf("TLS Error: certificateFile:[%v]  privateKeyFile:[%v] %v\n", certificateFile, privateKeyFile, err)
	}
	pem, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("TLS Error: [%v] %v\n\n", caFile, err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pem) {
		log.Fatal("TLS Error: Failed appending certs")
	}

	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0] = mycert
	config.RootCAs = certPool
	config.ClientCAs = certPool
	config.ClientAuth = tls.RequireAndVerifyClientCert

	config.CipherSuites = []uint16{
		// tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	}

	config.MinVersion = tls.VersionTLS12
	config.SessionTicketsDisabled = true
	//log.Println("BEFORE BuildNameToCertificate() : ")
	//log.Printf("Config.NameToCertificate :%v\n", config.NameToCertificate)
	config.BuildNameToCertificate()
	//log.Println("AFTER BuildNameToCertificate() : ")
	//log.Printf("Config.NameToCertificate :%v\n", config.NameToCertificate)
	config.PreferServerCipherSuites = true

	return config
}
