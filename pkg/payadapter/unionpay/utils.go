// utils.go
package unionpay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"errors"
)

// Pkcs8PrivateKey 解析 PKCS#8 私钥结构
type Pkcs8PrivateKey struct {
	Version    int
	Algo       []byte
	PrivateKey []byte
}

// LoadPrivateKeyFromPfx 从PFX/P12证书文件加载RSA私钥
// pfxData: PFX/P12证书文件内容
// password: 证书密码
func LoadPrivateKeyFromPfx(pfxData []byte, password string) (*rsa.PrivateKey, error) {
	// 首先尝试解析为PEM格式
	block, _ := pem.Decode(pfxData)
	if block != nil {
		// 如果是PEM格式，尝试解密
		if x509.IsEncryptedPEMBlock(block) {
			decryptedBlock, err := x509.DecryptPEMBlock(block, []byte(password))
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt PEM block: %w", err)
			}

			// 尝试解析PKCS#8私钥
			var pkcs8 Pkcs8PrivateKey
			_, err = asn1.Unmarshal(decryptedBlock, &pkcs8)
			if err == nil {
				// 成功解析PKCS#8结构
				key, err := x509.ParsePKCS8PrivateKey(pkcs8.PrivateKey)
				if err != nil {
					return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
				}

				rsaKey, ok := key.(*rsa.PrivateKey)
				if !ok {
					return nil, errors.New("not RSA private key")
				}
				return rsaKey, nil
			}

			// 如果不是PKCS#8，尝试解析PKCS#1
			key, err := x509.ParsePKCS1PrivateKey(decryptedBlock)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS#1 private key: %w", err)
			}
			return key, nil
		}

		// 如果是未加密的PEM块，直接解析
		if block.Type == "PRIVATE KEY" {
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
			}

			rsaKey, ok := key.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("not RSA private key")
			}
			return rsaKey, nil
		}

		if block.Type == "RSA PRIVATE KEY" {
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS#1 private key: %w", err)
			}
			return key, nil
		}
	}

	// 如果不是PEM格式，可能是二进制PKCS#12格式
	// 这里需要导入额外的库来解析PKCS#12，目前先返回错误
	return nil, errors.New("PKCS#12 format not supported directly, please convert to PEM format first")
}

// LoadPrivateKeyFromFile 从文件加载私钥，支持PEM格式
func LoadPrivateKeyFromFile(filePath, password string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	return LoadPrivateKeyFromPfx(data, password)
}

// LoadPublicKeyFromFile 从文件加载公钥
func LoadPublicKeyFromFile(filePath string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	if block.Type != "PUBLIC KEY" && block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected block type: %s", block.Type)
	}

	if block.Type == "PUBLIC KEY" {
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
		}

		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not RSA public key")
		}
		return rsaPub, nil
	}

	// RSA PUBLIC KEY 格式
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#1 public key: %w", err)
	}

	return pub, nil
}
