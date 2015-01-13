package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// Used by the RPC methods
type Args struct {
	ClientNmae   string
	PolicyName   string
	ScheduleName string
	ScheduleType string
	Status       string
	ResuleFile   string
}

// Use by RPC server and client for RPC status codes
type Results struct {
	Code int
}

// Struct for server and client configuration
type ConfigSettings struct {
	TLSCommonCA     string `json:"TLSCommonCA"`
	TLSMyCert       string `json:"TLSMyCert"`
	TLSMyKey        string `json:"TLSMyKey"`
	ServerIPPort    string `json:"ServerIPPort"`
	ServerIP        string `json:"ServerIP"`
	ServerPort      int    `json:"ServerPort"`
	ExternalCMDPath string `json:"ExternalCMDPath"`
	ExternalCMD     string `json:"ExternalCMD"`
	ClientIP        string `json:"ClientIP"`
}

// Parse given config file into struct
func ParseConfig(filename string) (c *ConfigSettings) {
	// log.Fatal(filename)
	configFile, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error opening configuration file: ", filename, " ", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&c); err != nil {
		log.Fatal("Error parsing configuration file: ", filename, " ", err.Error())
	}
	c.ServerIPPort = c.ServerIP + ":" + strconv.Itoa(c.ServerPort)
	log.Printf("Settings:")
	log.Printf("TLSCommonCA: %s\n", c.TLSCommonCA)
	log.Printf("TLSMyCert: %s\n", c.TLSMyCert)
	log.Printf("TLSMyKey: %s\n", c.TLSMyKey)
	log.Printf("ServerIPPort: %s\n", c.ServerIPPort)
	log.Printf("ServerIP: %s\n", c.ServerIP)
	log.Printf("ServerPort: %s\n", c.ServerPort)
	log.Printf("ExternalCMDPath: %s\n", c.ExternalCMDPath)
	log.Printf("ExternalCMD: %s\n", c.ExternalCMD)
	log.Printf("ClientIP: %s\n", c.ClientIP)
	return c
}

// Original is from github.com/hydrogen18/test-tls
func MustLoadCertificates(myCaCertificate, myCertificate, myPrivateKey string) (tls.Certificate, *x509.CertPool) {
	privateKeyFile := myPrivateKey
	certificateFile := myCertificate
	caFile := myCaCertificate

	mycert, err := tls.LoadX509KeyPair(certificateFile, privateKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	pem, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pem) {
		log.Fatal("Failed appending certs")
	}

	return mycert, certPool
}

func MustGetTlsConfiguration(myCaCertificate, myCertificate, myPrivateKey string) *tls.Config {
	config := &tls.Config{}
	mycert, certPool := MustLoadCertificates(myCaCertificate, myCertificate, myPrivateKey)
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0] = mycert

	config.RootCAs = certPool
	config.ClientCAs = certPool

	config.ClientAuth = tls.RequireAndVerifyClientCert

	//ordered as per https://www.grc.com/miscfiles/SChannel_Cipher_Suites.txt
	config.CipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	}

	//Use only TLS v1.2
	config.MinVersion = tls.VersionTLS12

	//Don't allow session resumption
	config.SessionTicketsDisabled = true
	return config
}
