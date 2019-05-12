package sdk

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/pjoc-team/base-service/pkg/logger"
)

func packageData(originalData []byte, packageSize int) (r [][]byte) {
	var src = make([]byte, len(originalData))
	copy(src, originalData)

	r = make([][]byte, 0)
	if len(src) <= packageSize {
		return append(r, src)
	}
	for len(src) > 0 {
		var p = src[:packageSize]
		r = append(r, p)
		src = src[packageSize:]
		if len(src) <= packageSize {
			r = append(r, src)
			break
		}
	}
	return r
}

func RSAEncrypt(plaintext, key []byte) ([]byte, error) {
	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return nil, errors.New("public key error")
	}

	var pubInterface interface{}
	pubInterface, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	var pub = pubInterface.(*rsa.PublicKey)

	var data = packageData(plaintext, pub.N.BitLen()/8-11)
	var cipherData []byte = make([]byte, 0, 0)

	for _, d := range data {
		var c, e = rsa.EncryptPKCS1v15(rand.Reader, pub, d)
		if e != nil {
			return nil, e
		}
		cipherData = append(cipherData, c...)
	}

	return cipherData, nil
}

func RSADecrypt(ciphertext, key []byte) ([]byte, error) {
	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return nil, errors.New("private key error")
	}

	var pri *rsa.PrivateKey
	pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	var data = packageData(ciphertext, pri.PublicKey.N.BitLen()/8)
	var plainData []byte = make([]byte, 0, 0)

	for _, d := range data {
		var p, e = rsa.DecryptPKCS1v15(rand.Reader, pri, d)
		if e != nil {
			return nil, e
		}
		plainData = append(plainData, p...)
	}
	return plainData, nil
}

func SignPKCS1v15(src, key []byte, hash crypto.Hash) ([]byte, error) {
	var h = hash.New()
	h.Write(src)
	var hashed = h.Sum(nil)

	var err error
	var block *pem.Block
	block, data := pem.Decode(key)
	var pri *rsa.PrivateKey
	if block == nil {
		pri, err = x509.ParsePKCS1PrivateKey(data)
		//return nil, errors.New("private key error")
	} else {
		pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	if err != nil {
		return nil, err
	}
	return rsa.SignPKCS1v15(rand.Reader, pri, hash, hashed)
}

func VerifyPKCS1v15(src, sig, key []byte, hash crypto.Hash) error {
	var h = hash.New()
	h.Write(src)
	var hashed = h.Sum(nil)

	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return errors.New("public key error")
	}

	var pubInterface interface{}
	pubInterface, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	var pub = pubInterface.(*rsa.PublicKey)

	return rsa.VerifyPKCS1v15(pub, hash, hashed, sig)
}

func SignPKCS8(src []byte, privateKey string, hash crypto.Hash) ([]byte, error) {
	var h = hash.New()
	h.Write(src)
	var hashed = h.Sum(nil)

	var err error
	bytes, _ := base64.StdEncoding.DecodeString(privateKey)
	//var block *pem.Block
	//block, _ = pem.Decode(privateKey)
	//if block == nil {
	//	return nil, errors.New("private key error")
	//}

	//var pri *rsa.PrivateKey
	pri, err := x509.ParsePKCS8PrivateKey(bytes)
	if err != nil {
		logger.Log.Errorf("Parse private key with error: %v", err.Error())
		return nil, err
	}
	//rsa.Sign
	return rsa.SignPKCS1v15(rand.Reader, pri.(*rsa.PrivateKey), hash, hashed)
}

func VerifyPKCS1v15WithStringKey(src, sig []byte, publicKeyString string, hash crypto.Hash) error {
	publicKey := ParsePublicKey(publicKeyString)
	return VerifyPKCS1v15(src, sig, publicKey, hash)
}

//func VerifyPKCS8(src []byte, sig, publicKey string, hash crypto.Hash) error {
//	var h = hash.New()
//	h.Write(src)
//	var hashed = h.Sum(nil)
//
//	var err error
//	bytes, _ := base64.StdEncoding.DecodeString(publicKey)
//	pubInterface, err = x509.ParsePKCS8PrivateKey(block.Bytes)
//	if err != nil {
//		return err
//	}
//	var pub = pubInterface.(*rsa.PublicKey)
//
//	return rsa.VerifyPKCS1v15(pub, hash, hashed, sig)
//}
