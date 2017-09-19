package http

import (
	"crypto/aes"
	"crypto/cipher"
	prng "crypto/rand"
	"encoding/base64"

	"github.com/gin-gonic/gin"
)

func InitCmdContext(c *gin.Context) {
	ctx := &cmdContext{
		Context: c,
		cmdID:   makeCmdID(),
	}
}

var sr cipher.Stream

func makeCmdID() string {
	if sr == nil {
		key := make([]byte, 16)
		prng.Read(key)
		aesCipher, _ := aes.NewCipher(key)
		sr = cipher.NewCTR(aesCipher)
	}

	b := make([]byte, 32)
	sr.XORKeyStream(b, b)

	return base64.StdEncoding.EncodeToString(b)
}

func GinHandler(func(*cmdContext)) gin.HandlerFunc {
}
