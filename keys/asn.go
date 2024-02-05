package keys

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	_ "unsafe"
)

// @see crypto/x509.publicKeyInfo
type pkixPublicKey struct {
	Raw       asn1.RawContent
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}

// @see crypto/x509.pkcs1PublicKey
type pkcs1PublicKey struct {
	N *big.Int
	E int
}

// @see crypto/x509.pkcs1PrivateKey
type pkcs1PrivateKey struct {
	Version int
	N       *big.Int
	E       int
	D       *big.Int
	P       *big.Int
	Q       *big.Int
	Dp      *big.Int `asn1:"optional"`
	Dq      *big.Int `asn1:"optional"`
	Qinv    *big.Int `asn1:"optional"`

	AdditionalPrimes []pkcs1AdditionalRSAPrime `asn1:"optional,omitempty"`
}

// @see crypto/x509.pkcs1AdditionalRSAPrime
type pkcs1AdditionalRSAPrime struct {
	Prime *big.Int

	Exp   *big.Int
	Coeff *big.Int
}

// @see crypto/x509.pkcs8
type pkcs8PrivateKey struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
}
