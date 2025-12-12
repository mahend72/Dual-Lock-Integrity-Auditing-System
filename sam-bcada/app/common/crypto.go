package common

import (
	cryptosha256
	encodinghex
)

 For real use load N and g from config  key setup
type RsaHvtParams struct {
	N string  hex
	G string  hex
}

func LoadRsaHvtParams() RsaHvtParams {
	return RsaHvtParams{
		N C7E5...,  big hex modulus
		G 05,       generator
	}
}

 HashHex returns SHA-256 hash hex of data
func HashHex(data []byte) string {
	h = sha256.Sum256(data)
	return hex.EncodeToString(h[])
}
