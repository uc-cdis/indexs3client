package handlers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash/crc32"
	"io"
	"sync"
)

type HashInfo struct {
	Crc32c string
	Md5    string
	Sha1   string
	Sha256 string
	Sha512 string
}

// CalculateBasicHashes generates hases
func CalculateBasicHashes(rd io.Reader) (*HashInfo, error) {
	crc32c := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	md5 := md5.New()
	sha1 := sha1.New()
	sha256 := sha256.New()
	sha512 := sha512.New()

	multiWriter := io.MultiWriter(crc32c, md5, sha1, sha256, sha512)

	if _, err := io.Copy(multiWriter, rd); err != nil {
		return nil, err
	}

	return &HashInfo{
		Crc32c: hex.EncodeToString(crc32c.Sum(nil)),
		Md5:    hex.EncodeToString(md5.Sum(nil)),
		Sha1:   hex.EncodeToString(sha1.Sum(nil)),
		Sha256: hex.EncodeToString(sha256.Sum(nil)),
		Sha512: hex.EncodeToString(sha512.Sum(nil)),
	}, nil

}

type multiCWriter struct {
	writers []io.Writer
}

func (t *multiCWriter) Write(p []byte) (n int, err error) {
	type data struct {
		n   int
		err error
	}

	results := make([]data, len(t.writers))

	var wg sync.WaitGroup
	for idx, w := range t.writers {
		wg.Add(1)
		go func(wr io.Writer, p []byte, res *data) {
			defer wg.Done()
			res.n, res.err = wr.Write(p)
			if res.n != len(p) {
				res.err = io.ErrShortWrite
			}
		}(w, p, &results[idx])
	}
	wg.Wait()
	for idx := range t.writers {
		if results[idx].err != nil {
			return results[idx].n, results[idx].err
		}
	}

	return len(p), nil
}
