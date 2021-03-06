package handlers

import (
	"encoding/hex"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestHash(t *testing.T) {
	input := "This is a test"
	//r := strings.NewReader(input)

	hashs := CreateNewHashCollection()
	hashs, _ = UpdateBasicHashes(hashs, []byte(input))

	//hashInfo, _ := CalculateBasicHashes(r)
	assert.Equal(t, hex.EncodeToString(hashs.Md5.Sum(nil)), "ce114e4501d2f4e2dcea3e17b546f339")
	assert.Equal(t, hex.EncodeToString(hashs.Crc32c.Sum(nil)), "d8ad940d")
	assert.Equal(t, hex.EncodeToString(hashs.Sha1.Sum(nil)), "a54d88e06612d820bc3be72877c74f257b561b19")
	assert.Equal(t, hex.EncodeToString(hashs.Sha256.Sum(nil)), "c7be1ed902fb8dd4d48997c6452f5d7e509fbcdbe2808b16bcf4edce4c07d14e")
	assert.Equal(t, hex.EncodeToString(hashs.Sha512.Sum(nil)), "a028d4f74b602ba45eb0a93c9a4677240dcf281a1a9322f183bd32f0bed82ec72de9c3957b2f4c9a1ccf7ed14f85d73498df38017e703d47ebb9f0b3bf116f69")
}
