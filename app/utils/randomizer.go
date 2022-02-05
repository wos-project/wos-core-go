package utils

import (
	"encoding/base64"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// GenerateBase64Rand generates a base64 random string
func GenerateBase64Rand() string {
	b1 := uuid.NewV4()
	b2 := uuid.NewV4()
	b3 := uuid.NewV4()
	return base64.RawURLEncoding.EncodeToString(b1.Bytes()) + base64.RawURLEncoding.EncodeToString(b2.Bytes()) + base64.RawURLEncoding.EncodeToString(b3.Bytes())
}

// GenerateRandomNumber generates a random number
func GenerateRandomNumber() int {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	return r1.Intn(1000000)
}

// Generate ipfs style cid
func GenerateCid() (string, error) {
	pref := cid.Prefix{
		Version: 1,
		Codec: cid.Raw,
		MhType: mh.SHA2_256,
		MhLength: -1, // default length
	}
	
	// And then feed it some data
	c, err := pref.Sum([]byte(GenerateBase64Rand()))
	if err != nil {
		return "", err
	}
	return c.String(), nil
}
