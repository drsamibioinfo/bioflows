package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/bioflows/src/bioflows/helpers"
	"io/ioutil"
	"strings"
)

func ReadX509Certificate(certPath string) (*x509.Certificate , error) {
	if !helpers.PathExists(certPath) {
		return nil , errors.New("Certificate file doesn't exist")
	}
	certData , err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil , errors.New("Unable to read certificate file from desk")
	}
	certBlock , _ := pem.Decode([]byte(certData))
	if certBlock == nil {
		return nil , errors.New("Unable to read certificate data blocks.")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil , errors.New("Unable to read certificate data blocks.")
	}
	return cert , nil
}
// This function will read private Key from a file path
func ReadPrivKeyFile(filePath string,password string) (*rsa.PrivateKey,error) {
	if !helpers.PathExists(filePath) {
		return nil , errors.New("Private Key file doesn't exist")
	}
	privKey, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil , errors.New("Private Key doesn't exist")
	}
	privPem , _ := pem.Decode(privKey)
	if !strings.Contains(privPem.Type,"PRIVATE KEY"){
		return nil , errors.New("Wrong Key loaded. This is not a private Key")
	}
	privPemBytes := privPem.Bytes
	if x509.IsEncryptedPEMBlock(privPem) {
		if len(password) < 1 {
			return nil, errors.New("Password is required to decode the encrypted private key.")
		}
		privPemBytes , err = x509.DecryptPEMBlock(privPem,[]byte(password))
	}
	var parsedKey interface{}
	if parsedKey , err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey , err = x509.ParsePKCS8PrivateKey(privPemBytes) ; err != nil {
			return nil , errors.New("Unknown Private Key format provided. Unable to use this key.")
		}
	}
	return parsedKey.(*rsa.PrivateKey) , nil
}



