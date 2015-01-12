// Original is from
// github.com/hydrogen18/test-tls
package main

import (
        "crypto/tls"
        "crypto/x509"
        "io/ioutil"
        "log"
        "encoding/json"
        "os"
)

type ConfigSettings struct {
        TLSCommonCA     string `json:"TLSCommonCA"`
        TLSMyCert       string `json:"TLSMyCert"`
        TLSMyKey        string `json:"TLSMyKey"`
        ServerIP        string `json:"ServerIP"`
        ServerPort      int    `json:"ServerPort"`
        ExternalCMDPath string `json:"ExternalCMDPath"`
        ExternalCMD     string `json:"ExternalCMD"`
        ClientIP        string `json:"ClientIP"`
}

func ParseConfig(filename string) (c *ConfigSettings) {
        configFile, err := os.Open(filename)
        if err != nil {
                log.Fatal("Error opening configuration file: ", err.Error())
        }

        jsonParser := json.NewDecoder(configFile)
        if err = jsonParser.Decode(&c); err != nil {
                log.Fatal("Error parsing configuration file: ", err.Error())
        }
        return c
}




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

        //Use only modern ciphers
        // ordered as per https://www.grc.com/miscfiles/SChannel_Cipher_Suites.txt
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
