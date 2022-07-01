package guiver

import (
	"bytes"
	"errors"
	"hash/fnv"
	"sync"
	"time"

	"log"
	"os"
	"strconv"
)

type (
	// Cache is the struct we use to scope different collections with different lifetimes for each entry
	Cache struct {
		lifetime        time.Duration
		currKv          map[uint64]*entry
		nextKv          map[uint64]*entry
		lock            *sync.Mutex
		currentTimeStep int64
	}

	entry struct {
		ts   int64
		data interface{}
	}
)

type wgonfigClient interface {
	Get(string) (string, error)
	Set(string, string) error
}

var (
	client wgonfigClient
	cache  = New(time.Minute * 1)

	defaultApp = ""

	// CacheMiss is returned if we don't have an entry for any given key
	CacheMiss = errors.New("ttl cache miss")
	// CacheExpired is returned if the entry exists but is too old
	CacheExpired = errors.New("ttl cache entry expired")
)

const (
	osConst = "os"
)

func init() {
	source, isSet := os.LookupEnv("WGONFIG_SOURCE")
	if !isSet {
		log.Println("app using wgonfig without source configured. defaulting to OS")
		source = osConst
	}

	switch source {
	case osConst:
		log.Println("wgonfig using environment vars")
		client = &osClient{}
	default:
		log.Fatalf("app using wgonfig has invalid source %s", source)
	}
}

/*
	internal getter for wgonfig. on misses we expect callers to set a default value
	attempts to read from cache then falls back to true source
*/
func get(app, key string) (string, error) {
	// in the case of tests/os we dont prepend /
	if app != "" {
		key = BString(app, "/", key)
	}
	res, op := cache.Get(key)

	switch op {
	case nil: // cache hit
		return res.(string), nil
	default: //expired or miss
		res, err := client.Get(key)
		if err != nil {
			return "", err
		}
		cache.Set(key, res)
		return res, nil
	}
}

/*
	Get alias' that default to DEFAULT_APP
*/

// GetAppInt returns the int form of the value of key or def if not set
func GetAppInt(app, key string, def int) (int, error) {
	s, e := get(app, key)
	if e == nil {
		return strconv.Atoi(s)
	}

	//cache.Set(key, strconv.Itoa(def))
	return def, e
}

// GetAppString returns the string form of the value of key or def if not set
func GetAppString(app, key, def string) (string, error) {
	s, e := get(app, key)
	if e == nil && s != "" {
		return s, nil
	}

	cache.Set(key, def)
	return def, e
}

/*
	Get alias' that default to DEFAULT_APP app
*/

// GetInt returns the int form of key or def if not set
func GetInt(key string, def int) (int, error) {
	return GetAppInt(defaultApp, key, def)
}

// GetString returns the string form of key or def if not set
func GetString(key string, def string) (string, error) {
	return GetAppString(defaultApp, key, def)
}

// BString fast string concatenation
func BString(strings ...string) string {
	var b bytes.Buffer
	for _, s := range strings {
		b.WriteString(s)
	}
	return b.String()
}

func New(d time.Duration) *Cache {
	c := &Cache{
		lock:     &sync.Mutex{},
		currKv:   map[uint64]*entry{},
		nextKv:   map[uint64]*entry{},
		lifetime: d,
	}

	c.currentTimeStep = c.getCurrentTimeStep()

	return c
}

func (c *Cache) getCurrentTimeStep() int64 {
	return now() / int64(c.lifetime)
}

func now() int64 {
	return time.Now().UTC().UnixNano()
}

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// Get takes a key string and then hashes it to an int64 to use as the entry key
// second string is the result of the operation if failure
func (c *Cache) Get(sKey string) (interface{}, error) {
	key := hash(sKey)
	c.lock.Lock()
	defer c.lock.Unlock()
	if r := c.currKv[key]; r == nil {
		return nil, CacheMiss
	} else if time.Duration(now()-r.ts) > c.lifetime {
		delete(c.currKv, key)
		return nil, CacheExpired
	} else {
		return r.data, nil
	}
}

// Set sets the string key's value to the value provided
func (c *Cache) Set(sKey string, value interface{}) {
	key := hash(sKey)
	currStep := c.getCurrentTimeStep()

	c.lock.Lock()
	defer c.lock.Unlock()

	e := &entry{
		ts:   now(),
		data: value,
	}

	if c.currentTimeStep != currStep {
		c.currKv = c.nextKv
		c.nextKv = map[uint64]*entry{}
		c.currentTimeStep = currStep
	}

	c.nextKv[key] = e
	c.currKv[key] = e
}
