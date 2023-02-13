package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"meipian.cn/meigo/v2/log"
)

// BuildQuery query 格式化
// 向前兼容
func BuildQuery(m map[string]interface{}) string {
	return parseQuery(m)
}

// query 格式化
func parseQuery(m map[string]interface{}) (query string) {
	values := parseQueryValues(m)
	return values.Encode()
}

// parseQueryValues 解析参数到url.Values
func parseQueryValues(m map[string]interface{}) url.Values {
	values := make(url.Values)

	for k, v := range m {
		item := reflect.ValueOf(v)
		switch item.Kind() {
		case reflect.Slice, reflect.Array:
			key := fmt.Sprintf("%s[]", k)
			for i := 0; i < item.Len(); i++ {
				rf := reflect.ValueOf(item.Index(i))
				if i == 0 {
					values.Set(key, fmt.Sprintf("%v", rf.Interface()))
				} else {
					values.Add(key, fmt.Sprintf("%v", rf.Interface()))
				}
			}
		default:
			values.Set(k, fmt.Sprintf("%v", v))
		}
	}
	return values
}

// appendUrlParam URL追加参数
func appendUrlParam(inputUrl string, m map[string]interface{}) (newUrl string) {
	if len(m) == 0 {
		return inputUrl
	}
	values := parseQueryValues(m)
	urlObj, err := url.Parse(inputUrl)
	if err != nil {
		return inputUrl
	}
	query := urlObj.Query()
	for key, value := range values {
		for _, v := range value {
			query.Add(key, v)
		}
	}
	urlObj.RawQuery = query.Encode()
	return urlObj.String()
}

// 一维数组格式化形式： map["name1"]["v1", "v2"]会格式化为name1[]=v1&name1[]=v2
func Get(url string, m map[string]interface{}) (ret []byte, code int) {
	url = appendUrlParam(url, m)

	defer func() {
		log.WithFields(log.Fields{
			"url":  url,
			"resp": string(ret),
		}).Debug("API请求：" + url)
	}()

	resp, err := http.Get(url)
	if err != nil {
		log.Err(err.Error())
		code = ErrAPI
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err.Error())
		code = ErrAPI
		return
	}

	return
}

func GetWithTimeOut(url string, m map[string]interface{}, timeout int64) (ret []byte, code int) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	url = appendUrlParam(url, m)

	startTime := time.Now()
	defer func() {
		log.WithFields(log.Fields{
			"url":      url,
			"timeout":  timeout,
			"duration": time.Since(startTime).Seconds(),
			"resp":     string(ret),
		}).Debug("API请求：" + url)
	}()

	resp, err := client.Get(url)
	if err != nil {
		log.Err(err.Error())
		code = ErrAPI
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err.Error())
		code = ErrAPI
		return
	}

	return
}

// Post Http Post Json；
func Post(url string, m map[string]interface{}) (ret []byte, code int) {
	return PostNew(url, m)
}

// / 带超时响应请求post
func PostWithTimeOut(url string, m map[string]interface{}, timeout int64) (ret []byte, err error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	body, err := json.Marshal(m)
	ret = []byte{}
	if err != nil {
		log.Err(err.Error())
		return nil, err
	}

	startTime := time.Now()
	defer func() {
		log.WithFields(log.Fields{
			"url":      url,
			"body":     string(body),
			"timeout":  timeout,
			"duration": time.Since(startTime).Seconds(),
			"resp":     string(ret),
		}).Debug("API请求：" + url)
	}()

	resp, err := client.Post(url, "application/json;", bytes.NewReader(body))
	if err != nil {
		log.Err(err.Error())
		return
	}
	errc := make(chan error, 1)
	go func() {
		d, err := ioutil.ReadAll(resp.Body)
		if len(d) > 0 {
			ret = d
		}
		errc <- err
		resp.Body.Close()
	}()

	const failTime = 5 * time.Second
	after := time.NewTimer(failTime)
	defer after.Stop()
	select {
	case err := <-errc:
		if err != nil {
			ne, ok := err.(net.Error)
			if !ok {
				log.Err(fmt.Sprintf("error value from ReadAll was %T; expected some net.Error", err))
			} else if !ne.Timeout() {
				log.Err("net.Error.Timeout = false; want true")
			}
			if got := ne.Error(); !strings.Contains(got, "Client.Timeout exceeded") {
				log.Err(fmt.Sprintf("error string = %q; missing timeout substring", got))
			}
		}
	// 之前使用 <-time.After() 会导致内存暴涨
	case <-after.C:
		log.Err(fmt.Sprintf("timeout after %v waiting for timeout of %v", failTime, timeout))
	}

	return
}

// PostNew Http Post Json；
func PostNew(url string, rq interface{}) (ret []byte, code int) {
	qb, err := json.Marshal(rq)
	if err != nil {
		log.Err(err.Error())
		code = ErrAPI
		return
	}
	ret, code = post(url, "application/json;", qb)
	return

}

func post(url string, contentType string, body []byte) (ret []byte, code int) {

	startTime := time.Now()
	defer func() {
		log.WithFields(log.Fields{
			"url":      url,
			"body":     string(body),
			"duration": time.Since(startTime).Seconds(),
			"resp":     string(ret),
		}).Debug("API请求：" + url)
	}()

	resp, err := http.Post(url, contentType, bytes.NewReader(body))
	if err != nil {
		log.Err(err.Error())
		code = ErrAPI
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err.Error())
		code = ErrAPI
		return
	}

	return
}

// PostFormU "application/x-www-form-urlencoded"
func PostFormU(url string, m map[string]interface{}) (ret []byte, code int) {

	body := parseQuery(m)

	ret, code = post(url, "application/x-www-form-urlencoded", []byte(body))
	return

}

func reqId() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
