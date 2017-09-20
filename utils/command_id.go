package utils

import (
	"crypto/aes"
	"crypto/cipher"
	prng "crypto/rand"
	"encoding/base64"
)

var sr cipher.Stream

func MakeCmdID() string {
	if sr == nil {
		key := make([]byte, 16)
		prng.Read(key)
		aesCipher, _ := aes.NewCipher(key)
		iv := make([]byte, 16)
		prng.Read(iv)
		sr = cipher.NewCTR(aesCipher, iv)
	}

	b := make([]byte, 32)
	sr.XORKeyStream(b, b)

	return base64.StdEncoding.EncodeToString(b)
}
