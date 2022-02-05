package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func TestRandomizer(t *testing.T) {
	cid, err := GenerateCid()
	assert.Nil(t, err)
	assert.Equal(t, 59, len(cid))
}

func TestCid(t *testing.T) {
	pref := cid.Prefix{
		Version: 1,
		Codec: cid.Raw,
		MhType: mh.SHA2_256,
		MhLength: -1, // default length
	}
	
	c, err := pref.Sum([]byte("hello world"))
	assert.Nil(t, err)
	assert.Equal(t, "bafkreifzjut3te2nhyekklss27nh3k72ysco7y32koao5eei66wof36n5e", c.String())
}

