package handlers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
)

type HashInfo struct {
	Crc32c string
	Md5    string
	Sha1   string
	Sha256 string
	Sha512 string
}

// CalculateBasicHashes generates hases
func CalculateBasicHashes(rd io.Reader) HashInfo {
	crc32c := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	md5 := md5.New()
	sha1 := sha1.New()
	sha256 := sha256.New()
	sha512 := sha512.New()

	multiWriter := MultiCWriter(crc32c, md5, sha1, sha256, sha512)

	_, err := io.Copy(multiWriter, rd)
	if err != nil {
		fmt.Println(err.Error())
	}

	var info HashInfo

	info.Crc32c = hex.EncodeToString(crc32c.Sum(nil))
	info.Md5 = hex.EncodeToString(md5.Sum(nil))
	info.Sha1 = hex.EncodeToString(sha1.Sum(nil))
	info.Sha256 = hex.EncodeToString(sha256.Sum(nil))
	info.Sha512 = hex.EncodeToString(sha512.Sum(nil))

	return info
}

type multiCWriter struct {
	writers []io.Writer
}

func (t *multiCWriter) Write(p []byte) (n int, err error) {
	type data struct {
		n   int
		err error
	}

	results := make(chan data)

	for _, w := range t.writers {
		go func(wr io.Writer, p []byte, ch chan data) {
			n, err = wr.Write(p)
			if err != nil {
				ch <- data{n, err}
				return
			}
			if n != len(p) {
				ch <- data{n, io.ErrShortWrite}
				return
			}
			ch <- data{n, nil} //completed ok
		}(w, p, results)
	}

	for range t.writers {
		d := <-results
		if d.err != nil {
			return d.n, d.err
		}
	}
	return len(p), nil
}

// MultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
func MultiCWriter(writers ...io.Writer) io.Writer {
	w := make([]io.Writer, len(writers))
	copy(w, writers)
	return &multiCWriter{w}
}
