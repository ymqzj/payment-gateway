package unionpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"sort"
)

// signature.go
func GenerateSign(params map[string]string, privateKey *rsa.PrivateKey) (string, error) {
	// 1. 过滤空值 + 排序
	var keys []string
	for k := range params {
		if params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 2. 拼接成 key1=value1&key2=value2...
	var signStr string
	for _, k := range keys {
		signStr += k + "=" + params[k] + "&"
	}
	signStr = signStr[:len(signStr)-1] // 去掉最后一个 &

	// 3. SHA256 + RSA 签名
	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest)
	if err != nil {
		return "", err
	}

	// 4. Base64 编码
	return base64.StdEncoding.EncodeToString(signature), nil
}
func VerifySign(params map[string]string, signature string, publicKey *rsa.PublicKey) bool {
	// 同样构造待验签字符串
	var keys []string
	for k := range params {
		if k != "signature" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var signStr string
	for _, k := range keys {
		signStr += k + "=" + params[k] + "&"
	}
	signStr = signStr[:len(signStr)-1]

	// 解码签名
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	// 验证
	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest, sigBytes)
	return err == nil
}
