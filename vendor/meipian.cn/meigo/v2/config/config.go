package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

var configs = struct {
	m map[interface{}]interface{}
	sync.RWMutex
}{m: make(map[interface{}]interface{})}

const configFile = ".yml"

func init() {
	dir, err := os.Getwd()
	path := configFile
	for {
		path = filepath.Join(dir, configFile)
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			if dir == "" {
				// panic("yaml config is not exist!")
				break
			}
			dir = dir[:strings.LastIndex(dir, "/")]
			continue
		} else {
			break
		}
	}
	bs, err := ioutil.ReadFile(path)
	if err == nil {
		err = yaml.Unmarshal(bs, &configs.m)
		if err != nil {
			panic("meigo init err: yaml.Unmarshal, " + err.Error())
		}
	}
}

// ReadFromFile 指定文件读取配置
// f 文件路径 例:/Users/lanjingren/go/src/meipian.cn/meigo/v2/config/.yml2
func ReadFromFile(f string) (err error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(bs, &m)
	if err != nil {
		return
	}

	for k, v := range m {
		configs.m[k] = v
	}
	return
}

func Get(key interface{}) (val interface{}) {
	configs.RLock()
	defer configs.RUnlock()

	val, _ = configs.m[key]
	return
}

func GetStr(key string) (val string) {
	configs.RLock()
	defer configs.RUnlock()

	data, ok := configs.m[key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", data)
}

func Set(key, value interface{}) {
	configs.Lock()
	defer configs.Unlock()
	configs.m[key] = value
}

func GetStringDft(key, dft string) (val string) {
	data := GetStr(key)
	if data == "" {
		return dft
	}

	val = data
	return
}

func GetDft(key, dft string) (val string) {
	data := GetStr(key)
	if data == "" {
		return dft
	}

	val = data
	return
}

func GetInt(key string) (val int) {
	data := GetStr(key)
	val, _ = strconv.Atoi(data)
	return
}

func GetIntDft(key string, dft int) (val int) {
	data := GetStr(key)
	if data == "" {
		return dft
	}

	val, _ = strconv.Atoi(data)
	return
}

func GetBool(key string, dft bool) (val bool) {
	data := GetStr(key)
	if data == "" {
		return dft
	}

	val, _ = strconv.ParseBool(data)
	return
}
