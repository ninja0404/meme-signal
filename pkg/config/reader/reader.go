package reader

import (
	"time"

	"github.com/ninja0404/meme-signal/pkg/config/source"
)

// Reader is an interface for merging changesets
type Reader interface {
	Merge(...*source.ChangeSet) (*source.ChangeSet, error)
	Values(*source.ChangeSet) (Values, error)
	String() string
}

// Values is returned by the reader
type Values interface {
	Bytes() []byte
	Get(path ...string) Value
	Set(val interface{}, path ...string)
	Del(path ...string)
	Map() map[string]interface{}
	Scan(v interface{}) error
}

// Value represents a value of any type
type Value interface {
	Bool(def bool) bool
	Int(def int) int
	Int64(def int64) int64
	String(def string) string
	Float64(def float64) float64
	Duration(def time.Duration) time.Duration
	StringSlice(def []string) []string
	StringMap(def map[string]string) map[string]string
	Scan(val interface{}) error
	Bytes() []byte
}
