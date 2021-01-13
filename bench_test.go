package hashcomp_test

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash/fnv"
	"math/rand"
	"testing"

	"github.com/sebnyberg/hashcomp"
	"golang.org/x/crypto/sha3"
)

var res []byte

func benchmarkHash(b *testing.B, hashlen int, h hashcomp.Hasher) {
	data := make([]byte, hashlen)
	b.ResetTimer()
	var hashResult []byte
	for i := 0; i < b.N; i++ {
		rand.Read(data)
		hashResult = h.Sum(data)
		h.Reset()
	}
	res = hashResult
}

func BenchmarkHash(b *testing.B) {
	for _, tc := range []struct {
		h    hashcomp.Hasher
		name string
	}{
		{fnv.New32(), "fnv1-32"},
		{fnv.New32a(), "fnv1a-32"},
		{fnv.New64(), "fnv1-64"},
		{fnv.New64a(), "fnv1a-64"},
		{fnv.New128(), "fnv1-128"},
		{fnv.New128a(), "fnv1a-128"},
		{&hashcomp.XXH3{}, "xxh3-64"},
		{&hashcomp.XXH3128{}, "xxh3-128"},
		{md5.New(), "md5"},
		{sha1.New(), "sha-1"},
		{sha256.New(), "sha-256"},
		{sha512.New(), "sha-512"},
		{sha3.New256(), "sha3-256"},
		{sha3.New512(), "sha3-512"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			benchmarkHash(b, 24, tc.h)
		})
	}
}
