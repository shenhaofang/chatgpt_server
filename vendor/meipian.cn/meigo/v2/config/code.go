package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var codes = map[int]string{}

const codeFile = "code.yml" // code 随版本分发

// init 初始化codes
func init() {
	bs, _ := ioutil.ReadFile(codeFile) // 无code, 不报错

	err := yaml.Unmarshal(bs, &codes)
	if err != nil {
		panic("meigo config: initCodes err: " + err.Error())
	}
	return
}

// CodeMsg 通过code取message
func CodeMsg(code int) (msg string) {
	msg, ok := codes[code]
	if !ok {
		return ""
	}
	return
}
