# Hash comparison

Quick comparison of some hash functions.

This is not an attempt at correctness. Please refer to SMHasher or the xxhash benchmarks for more solid methodology.


## Benchmarks 

```bash
$ go test -bench=. 
goos: darwin
goarch: amd64
pkg: github.com/sebnyberg/hashcomp
BenchmarkHash/fnv1-32-6                 14948162                79.2 ns/op
BenchmarkHash/fnv1a-32-6                15052020                80.7 ns/op
BenchmarkHash/fnv1-64-6                 14773922                79.0 ns/op
BenchmarkHash/fnv1a-64-6                15231482                79.2 ns/op
BenchmarkHash/fnv1-128-6                15102949                79.7 ns/op
BenchmarkHash/fnv1a-128-6               14862602                80.7 ns/op
BenchmarkHash/xxh3-64-6                 20052289                60.5 ns/op
BenchmarkHash/xxh3-128-6                17789229                67.3 ns/op
BenchmarkHash/md5-6                      6902458               174 ns/op
BenchmarkHash/sha-1-6                    5836796               206 ns/op
BenchmarkHash/sha-256-6                  4334325               276 ns/op
BenchmarkHash/sha-512-6                  3497335               345 ns/op
BenchmarkHash/sha3-256-6                 1670280               718 ns/op
BenchmarkHash/sha3-512-6                 1753620               684 ns/op
```

## Collision resistance

The expected collision resistance of a perfectly random hash can be derived mathematically (see [this article](https://www.ilikebigbits.com/2018_10_20_estimating_hash_collisions.html#:~:text=Whatever%20the%20usage%20you%20should,service%20attacks%2C%20spoofing%20and%20worse.))

| Hash space (H) | # elements (n) | Expected number of collisions (P) |
| -------------- | -------------- | --------------------------------- |
| 32bit (2^32)   | 10K            | 0.0116                            |
| 32bit (2^32)   | 100K           | 1.16                              |
| 32bit (2^32)   | 1M             | 116.41                            |
| 32bit (2^32)   | 10M            | 11641.53                          |
| 64bit (2^64)   | 100M           | 0                                 |
| 64bit (2^64)   | 1B             | 0.03                              |
| 64bit (2^64)   | 10B            | 2.71                              |
| 128bit (2^128) | 1e17           | 0                                 |
| 128bit (2^128) | 1e18           | 0.001469                          |
| 128bit (2^128) | 1e19           | 0.1469                            |
| 128bit (2^128) | 1e20           | 14.69                             |

This is what we expect to see from the practical collision test.

MurmurHash3 was excluded from the tests because it does not map to the same outputs on different CPU architectures. In my view, this makes alternative hashing methodologies more attractive (FNV-1 or xxhash).

The two datasets used for testing were:

1. English vocabulary (found in words.txt)
2. Random alpha-numerical strings from 8-32 bytes in length

### Collisions for 32-bit output hashes

In the example below, the fnv-1a 64 and 128bit output was sliced to only include the first 32 bits. This is obviously not how fnv-1a should be used, but it does tell us something about the properties of the hash. It is not uniformly random across bits, but rather random in the context of the entire hash.

It does not seem like the choice of hash method for 32-bit matters in terms of collisions (unless you slice something in a way you're not supposed to). Knowing how the number of elements relate to the output length is much more important. In this case, it appears that 32-bit is fine up to around 50K hashes, and already at 100K some hashes see collisions (SHA-512, SHA3-256).

```
=================================================
Collisions when output := hash[:4] 
-------------------------------------------------
\               fnv1a-32 fnv1a-64 fnv1a-128 sha-1   sha-256 sha-512 sha3-256 sha3-512 md5
EngWords        23       24       44025     16      25      27      32       28       18
Rand,100K,8-32B 0        0        3259      0       0       2       2        0        0
Rand,1M,8-32B   106      111      33865     126     124     113     107      136      133
Rand,10M,8-32B  11619    11529    350410    11551   11726   11486   11501    11728    11626
Rand,100M,8-32B 1154328  1157670  4581028   1157848 1156614 1154764 1154288  1155791  1155261
```

### Collisions for 64-bit hashes

Now slicing to 64-bits, I could not get a single collision (except for the misuse of FNV-1a 128bit). Testing a bigger input would require storing the output on disk rather than RAM, so I stopped at around 100M.

```bash
=================================================
Collisions when output := hash[:8] 
-------------------------------------------------
\               fnv1a-32 fnv1a-64 fnv1a-128 sha-1 sha-256 md5 xxh3
EngWords        23       0        0         0     0       0   0
Rand,100K,8-32B 0        0        0         0     0       0   0
Rand,1M,8-32B   106      0        0         0     0       0   0
Rand,10M,8-32B  11619    0        2         0     0       0   0
Rand,100M,8-32B 1154160  0        57        0     0       0   0
=================================================
```

