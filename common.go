// Original is from
// github.com/hydrogen18/test-tls
package main

import (
	"crypto/tls"
	"crypto/x509"
)

func MustLoadCertificates(myCaCertificate, myCertificate, myPrivateKey string) (tls.Certificate, *x509.CertPool) {

	mycert, err := tls.X509KeyPair([]byte(myCertificate), []byte(myPrivateKey))
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(myCaCertificate)) {
		panic("Failed appending certs")
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
