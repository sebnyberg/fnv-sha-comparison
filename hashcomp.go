package hashcomp

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"text/tabwriter"

	"github.com/gofrs/uuid"
	"github.com/zeebo/xxh3"
)

type Hasher interface {
	io.Writer
	Sum(b []byte) []byte
	Reset()
}

type XXH3128 struct {
	res [2]uint64
}

func (h *XXH3128) Write(s []byte) (n int, err error) {
	h.res = xxh3.Hash128(s)
	return len(s), nil
}

func (h *XXH3128) Reset() {}

func (h *XXH3128) Sum(b []byte) []byte {
	uintBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(uintBytes[:8], h.res[0])
	binary.BigEndian.PutUint64(uintBytes[8:], h.res[1])
	return uintBytes
}

type XXH3 struct {
	res uint64
}

func (h *XXH3) Write(s []byte) (n int, err error) {
	h.res = xxh3.Hash(s)
	return len(s), nil
}

func (h *XXH3) Reset() {}

func (h *XXH3) Sum(b []byte) []byte {
	uintBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(uintBytes, h.res)
	return uintBytes
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func GetUUIDs(n int) []string {
	keyMap := make(map[string]struct{})
	keys := make([]string, 0)
	for i := 0; i < n; i++ {
		u, _ := uuid.NewV4()
		k := u.String()
		if _, exists := keyMap[k]; exists {
			i--
			continue
		}
		keyMap[k] = struct{}{}
		keys = append(keys, k)
	}
	return keys
}

func GetWords() []string {
	keyMap := make(map[string]struct{})
	keys := make([]string, 0)
	f, err := os.Open("words.txt")
	check(err)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k := sc.Text()
		if _, exists := keyMap[k]; exists {
			log.Fatalln("duplicate word", k)
		}
		keyMap[k] = struct{}{}
		keys = append(keys, k)
	}
	return keys
}

func GetRandom(n int, strlen int) []string {
	keyMap := make(map[string]struct{})
	keys := make([]string, 0)
	for i := 0; i < n; {
		k := RandStringBytes(strlen)
		if _, exists := keyMap[k]; exists {
			i--
			continue
		}
		keyMap[k] = struct{}{}
		keys = append(keys, k)
		i++
	}
	return keys
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func GetRandomVarLen(n int, minlen, maxlen int) []string {
	keyMap := make(map[string]struct{})
	keys := make([]string, 0)
	for i := 0; i < n; {
		l := min(minlen, rand.Intn(maxlen))
		k := RandStringBytes(l)
		if _, exists := keyMap[k]; exists {
			continue
		}
		keyMap[k] = struct{}{}
		keys = append(keys, k)
		i++
	}
	return keys
}

func CheckAll() {
	type dataset struct {
		name       string
		getItemsFn func() []string
	}
	datasets := []dataset{
		{"EngWords", GetWords},
		{"Rand,100K,8-32B", func() []string { return GetRandomVarLen(int(1e5), 8, 32) }},
		{"Rand,1M,8-32B", func() []string { return GetRandomVarLen(int(1e6), 8, 32) }},
		{"Rand,10M,8-32B", func() []string { return GetRandomVarLen(int(1e7), 8, 32) }},
		{"Rand,100M,8-32B", func() []string { return GetRandomVarLen(int(1e8), 8, 32) }},
		// {"Rand,500M,8-32B", func() []string { return GetRandomVarLen(int(5e8), 8, 32) }},
		// {"Rand,100M,4-32B", func() []string { return GetRandomVarLen(int(1e8), 32) }},
		// {"Rand,10M,32B", func() []string { return GetRandom(int(1e7), 32) }},
		// {"Rand,100M,32B", func() []string { return GetRandom(int(1e8), 32) }},
		// {"UUID,5M,32B", func() []string { return GetUUIDs(int(5e6)) }},
	}

	type method struct {
		name string
		h    Hasher
	}
	hashMethods := []method{
		{"fnv1a-32", fnv.New32a()},
		{"fnv1a-64", fnv.New64a()},
		{"fnv1a-128", fnv.New128a()},
		{"sha-1", sha1.New()},
		{"sha-256", sha256.New()},
		// {"sha-512", sha512.New()},
		// {"sha3-256", sha3.New256()},
		// {"sha3-512", sha3.New512()},
		{"md5", md5.New()},
		{"xxh3-64", &XXH3{}},
		{"xxh3-128", &XXH3128{}},
	}

	maxlen := 8
	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 1, ' ', 0)
	fmt.Println("=================================================")
	fmt.Printf("Collisions when k := hash[:4] \n")
	fmt.Println("-------------------------------------------------")
	fmt.Fprint(tw, "\\")
	for _, hashMethod := range hashMethods {
		fmt.Fprintf(tw, "\t%v", hashMethod.name)
	}
	fmt.Fprintf(tw, "\n")
	for _, dataset := range datasets {
		fmt.Fprintf(os.Stderr, "creating dataset %v\n", dataset.name)
		fmt.Fprintf(tw, "%v", dataset.name)
		keys := dataset.getItemsFn()
		fmt.Fprintln(os.Stderr, "done!")
		for _, hashMethod := range hashMethods {
			fmt.Fprintf(os.Stderr, "running method %v\n", hashMethod.name)
			hashMethod.h.Reset()
			collisions := testHash(keys, hashMethod.h, maxlen)
			fmt.Fprintf(tw, "\t%v", collisions)
			fmt.Fprintln(os.Stderr, "done!")
		}
		fmt.Fprintf(tw, "\n")
	}
	tw.Flush()
	fmt.Println("=================================================")
}

func testHash(keys []string, h Hasher, maxlen int) (ncollisions int) {
	// defer func(start time.Time) {
	// 	fmt.Printf("Function execution took %v\n", time.Since(start))
	// }(time.Now())
	hashedKeys := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		h.Write([]byte(k))
		hashKey := string(h.Sum(nil))
		if len(hashKey) >= maxlen {
			hashKey = hashKey[:maxlen]
		}
		if _, exists := hashedKeys[hashKey]; exists {
			ncollisions++
		}
		hashedKeys[hashKey] = struct{}{}
		h.Reset()
	}
	return ncollisions
}

func calcP(nhashes float64, nbits int) float64 {
	return math.Pow(nhashes, 2) / (2 * float64(math.Pow(2, float64(nbits))))
}
