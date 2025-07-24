package json

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	simple "github.com/bitly/go-simplejson"
	"gopkg.in/yaml.v3"

	"github.com/ninja0404/meme-signal/pkg/config/reader"
	"github.com/ninja0404/meme-signal/pkg/config/source"
)

type jsonValues struct {
	ch *source.ChangeSet
	sj *simple.Json
}

type jsonValue struct {
	*simple.Json
}

func newValues(ch *source.ChangeSet) (reader.Values, error) {
	sj := simple.New()
	data, _ := ReplaceEnvVars(ch.Data)

	// 尝试直接解析JSON
	if err := sj.UnmarshalJSON(data); err != nil {
		// 如果是YAML格式，尝试先解析为map再转换为JSON
		if ch.Format == "yaml" {
			var yamlData map[string]interface{}
			if yamlErr := yaml.Unmarshal(data, &yamlData); yamlErr == nil {
				// 将yaml数据转换为JSON
				if jsonData, jsonErr := json.Marshal(yamlData); jsonErr == nil {
					if jsonParseErr := sj.UnmarshalJSON(jsonData); jsonParseErr == nil {
						return &jsonValues{ch, sj}, nil
					}
				}
			}
		}
		// 如果所有解析都失败，才设置为字符串
		sj.SetPath(nil, string(ch.Data))
	}
	return &jsonValues{ch, sj}, nil
}

func (j *jsonValues) Get(path ...string) reader.Value {
	return &jsonValue{j.sj.GetPath(path...)}
}

func (j *jsonValues) Del(path ...string) {
	// delete the tree?
	if len(path) == 0 {
		j.sj = simple.New()
		return
	}

	if len(path) == 1 {
		j.sj.Del(path[0])
		return
	}

	vals := j.sj.GetPath(path[:len(path)-1]...)
	vals.Del(path[len(path)-1])
	j.sj.SetPath(path[:len(path)-1], vals.Interface())
	return
}

func (j *jsonValues) Set(val interface{}, path ...string) {
	j.sj.SetPath(path, val)
}

func (j *jsonValues) Bytes() []byte {
	b, _ := j.sj.MarshalJSON()
	return b
}

func (j *jsonValues) Map() map[string]interface{} {
	m, _ := j.sj.Map()
	return m
}

func (j *jsonValues) Scan(v interface{}) error {
	b, err := j.sj.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (j *jsonValues) String() string {
	return "json"
}

func (j *jsonValue) Bool(def bool) bool {
	b, err := j.Json.Bool()
	if err == nil {
		return b
	}

	str, ok := j.Interface().(string)
	if !ok {
		return def
	}

	b, err = strconv.ParseBool(str)
	if err != nil {
		return def
	}

	return b
}

func (j *jsonValue) Int(def int) int {
	i, err := j.Json.Int()
	if err == nil {
		return i
	}

	str, ok := j.Interface().(string)
	if !ok {
		return def
	}

	i, err = strconv.Atoi(str)
	if err != nil {
		return def
	}

	return i
}

func (j *jsonValue) Int64(def int64) int64 {
	i, err := j.Json.Int64()
	if err == nil {
		return def
	}

	str, ok := j.Interface().(string)
	if !ok {
		return def
	}

	i, err = strconv.ParseInt(str, 10, 64) // 转换为 int64
	if err != nil {
		return def
	}

	return i
}

func (j *jsonValue) String(def string) string {
	return j.Json.MustString(def)
}

func (j *jsonValue) Float64(def float64) float64 {
	f, err := j.Json.Float64()
	if err == nil {
		return f
	}

	str, ok := j.Interface().(string)
	if !ok {
		return def
	}

	f, err = strconv.ParseFloat(str, 64)
	if err != nil {
		return def
	}

	return f
}

func (j *jsonValue) Duration(def time.Duration) time.Duration {
	v, err := j.Json.String()
	if err != nil {
		return def
	}

	value, err := time.ParseDuration(v)
	if err != nil {
		return def
	}

	return value
}

func (j *jsonValue) StringSlice(def []string) []string {
	v, err := j.Json.String()
	if err == nil {
		sl := strings.Split(v, ",")
		if len(sl) > 1 {
			return sl
		}
	}
	return j.Json.MustStringArray(def)
}

func (j *jsonValue) StringMap(def map[string]string) map[string]string {
	m, err := j.Json.Map()
	if err != nil {
		return def
	}

	res := map[string]string{}

	for k, v := range m {
		res[k] = fmt.Sprintf("%v", v)
	}

	return res
}

func (j *jsonValue) Scan(v interface{}) error {
	b, err := j.Json.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (j *jsonValue) Bytes() []byte {
	b, err := j.Json.Bytes()
	if err != nil {
		// try return marshalled
		b, err = j.Json.MarshalJSON()
		if err != nil {
			return []byte{}
		}
		return b
	}
	return b
}
