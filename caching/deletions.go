package caching

import "github.com/dgraph-io/ristretto"

var deletionsCache *ristretto.Cache // Connection ID -> IP

func setupDeletionsCache() {

	var err error
	deletionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // 1 million expected connections
		MaxCost:     1 << 30, // 1 GB
		BufferItems: 64,
	})

	if err != nil {
		panic(err)
	}
}

func AddForDeletion(id string, ip string) {
	deletionsCache.Set(id, ip, 1)
	deletionsCache.Wait()
}

func GetIP(id string) (string, bool) {

	ip, valid := deletionsCache.Get(id)
	if !valid {
		return "", false
	}

	return ip.(string), true
}
