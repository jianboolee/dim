package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrPrivateKeyNotFound = errors.New("私钥文件未找到")
	ErrPublicKeyNotFound  = errors.New("公钥文件未找到")

	ErrPrivateKeyLoaderError = errors.New("私钥加载失败")
	ErrPublicKeyLoaderError  = errors.New("公钥加载失败")

	ErrPrivateKeyParseFailed = errors.New("私钥解析失败")
	ErrPublicKeyParseFailed  = errors.New("公钥解析失败")

	ErrInvalidPEMBlock = errors.New("无效的PEM块")
)

// LoadRSAPrivateKey 从 PEM 文件加载 RSA 私钥，支持 PKCS#1 和 PKCS#8 格式
func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPrivateKeyLoaderError, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	var key *rsa.PrivateKey

	switch block.Type {
	case "RSA PRIVATE KEY": // PKCS#1 格式
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrPrivateKeyParseFailed, err)
		}
	case "PRIVATE KEY": // PKCS#8 格式
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrPrivateKeyParseFailed, err)
		}
		var ok bool
		key, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("%w: %w", ErrPrivateKeyParseFailed, err)
		}
	default:
		return nil, fmt.Errorf("%w: %w", ErrPrivateKeyParseFailed, err)
	}

	return key, nil
}

// LoadRSAPublicKey 从 PEM 文件加载 RSA 公钥，支持 PKIX 格式（PUBLIC KEY）和传统 RSA 格式（RSA PUBLIC KEY）
func LoadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPublicKeyLoaderError, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	var pubKey *rsa.PublicKey

	switch block.Type {
	case "PUBLIC KEY": // PKIX 格式（标准格式，如 openssl rsa -pubout 生成的）
		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrPublicKeyParseFailed, err)
		}
		var ok bool
		pubKey, ok = pubInterface.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("%w: %w", ErrPublicKeyParseFailed, err)
		}
	case "RSA PUBLIC KEY": // 传统 PKCS#1 格式（如 openssl rsa -RSAPublicKey_out 生成的）
		pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrPublicKeyParseFailed, err)
		}
	default:
		return nil, fmt.Errorf("%w: %w", ErrPublicKeyParseFailed, err)
	}

	return pubKey, nil
}

func LoadRSAPublicKeyFromPEMString(pemStr string) (*rsa.PublicKey, error) {
	publickey := strings.ReplaceAll(pemStr, `\n`, "\n")

	block, _ := pem.Decode([]byte(publickey))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return rsaPub, nil
}
