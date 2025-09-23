// utils.go
package unionpay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"

	"errors"
)

// Pkcs8PrivateKey 解析 PKCS#8 私钥结构
type Pkcs8PrivateKey struct {
	Version    int
	Algo       []byte
	PrivateKey []byte
}

func LoadPrivateKeyFromPfx(pfxData []byte, password string) (*rsa.PrivateKey, error) {
	// 首先尝试解析PFX/P12格式
	privateKey, err := x509.DecryptPEMBlock(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pfxData,
	}, []byte(password))
	if err != nil {
		// 如果不是PEM格式，尝试直接解析PKCS#12
		privateKey = pfxData
	}

	var pkcs8 Pkcs8PrivateKey
	_, err = asn1.Unmarshal(privateKey, &pkcs8)
	if err != nil {
		return nil, err
	}

	der := pkcs8.PrivateKey
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not rsa private key")
	}
	return rsaKey, nil
}
