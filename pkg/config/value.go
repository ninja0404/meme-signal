package config

import (
	"time"

	"github.com/ninja0404/meme-signal/pkg/config/reader"
)

type value struct{}

func newValue() reader.Value {
	return new(value)
}

func (v *value) Bool(def bool) bool {
	return false
}

func (v *value) Int(def int) int {
	return def
}

func (v *value) Int64(def int64) int64 {
	return def
}

func (v *value) String(def string) string {
	return ""
}

func (v *value) Float64(def float64) float64 {
	return 0.0
}

func (v *value) Duration(def time.Duration) time.Duration {
	return time.Duration(0)
}

func (v *value) StringSlice(def []string) []string {
	return nil
}

func (v *value) StringMap(def map[string]string) map[string]string {
	return map[string]string{}
}

func (v *value) Scan(val interface{}) error {
	return nil
}

func (v *value) Bytes() []byte {
	return nil
}
