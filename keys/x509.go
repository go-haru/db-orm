package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

const (
	PEMTypePKXIPublicKey  = "PUBLIC KEY"
	PEMTypePKCS1PublicKey = "RSA PUBLIC KEY"

	PEMTypePKCS8PrivateKey = "PRIVATE KEY"
	PEMTypePKCS1PrivateKey = "RSA PRIVATE KEY"
)

var ErrInvalidDataFormat = errors.New("invalid data format")

func errorOf[T any](_ T, err error) error { return err }

func typeAssertWithErr[T any](src any, err error) (dst T, _ error) {
	if err != nil {
		return dst, err
	}
	var ok bool
	if dst, ok = src.(T); !ok {
		return dst, fmt.Errorf("%w: expected: %T, got: %T", ErrInvalidDataFormat, dst, src)
	}
	return dst, nil
}

func ParseRSAPublicKeyDER(der []byte) (key *rsa.PublicKey, err error) {
	switch {
	case errorOf(asn1.Unmarshal(der, &pkixPublicKey{})) == nil:
		return typeAssertWithErr[*rsa.PublicKey](x509.ParsePKIXPublicKey(der))
	case errorOf(asn1.Unmarshal(der, &pkcs1PublicKey{})) == nil:
		return x509.ParsePKCS1PublicKey(der)
	default:
		return nil, ErrInvalidDataFormat
	}
}

func ParseRSAPublicKeyPEM(raw []byte) (key *rsa.PublicKey, err error) {
	var block *pem.Block
	if block, _ = pem.Decode(raw); block == nil {
		return nil, fmt.Errorf("%w: ref: %s", ErrInvalidDataFormat, "pem.type")
	}
	switch strings.ToUpper(block.Type) {
	case PEMTypePKXIPublicKey:
		return typeAssertWithErr[*rsa.PublicKey](x509.ParsePKIXPublicKey(block.Bytes))
	case PEMTypePKCS1PublicKey:
		return x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		return nil, fmt.Errorf("%w: ref: %s, expected: %T, got: %T", ErrInvalidDataFormat, "pem.type", block.Type, "[RSA ]PUBLIC KEY")
	}
}

func ParseRSAPrivateKeyDER(der []byte) (key *rsa.PrivateKey, err error) {
	switch {
	case errorOf(asn1.Unmarshal(der, &pkcs8PrivateKey{})) == nil:
		return typeAssertWithErr[*rsa.PrivateKey](x509.ParsePKCS8PrivateKey(der))
	case errorOf(asn1.Unmarshal(der, &pkcs1PrivateKey{})) == nil:
		return x509.ParsePKCS1PrivateKey(der)
	default:
		return nil, ErrInvalidDataFormat
	}
}

func ParseRSAPrivateKeyPEM(raw []byte) (key *rsa.PrivateKey, err error) {
	var block *pem.Block
	if block, _ = pem.Decode(raw); block == nil {
		return nil, fmt.Errorf("%w: ref: %s", ErrInvalidDataFormat, "pem.type")
	}
	switch strings.ToUpper(block.Type) {
	case PEMTypePKCS8PrivateKey:
		return typeAssertWithErr[*rsa.PrivateKey](x509.ParsePKCS8PrivateKey(block.Bytes))
	case PEMTypePKCS1PrivateKey:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("%w: ref: %s, expected: %T, got: %T", ErrInvalidDataFormat, "pem.type", block.Type, "[RSA ]PRIVATE KEY")
	}
}
