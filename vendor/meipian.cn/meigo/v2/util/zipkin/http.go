package zipkin

import (
	"bytes"
	"context"
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
	"meipian.cn/meigo/v2/util"
)

// RequestDefaultTimeout 请求默认超时时间
const RequestDefaultTimeout = 5
const Referer = "https://api.meipian.cn"

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
func Get(ctx context.Context, url string, m map[string]interface{}) (ret []byte, code int) {
	url = appendUrlParam(url, m)

	startTime := time.Now()
	defer func() {
		log.WithCtxFields(ctx, log.Fields{
			"url":      url,
			"duration": time.Since(startTime).Seconds(),
			"resp":     string(ret),
		}).Debug("API请求：" + url)
	}()

	req, _ := http.NewRequest("GET", url, nil)
	req = req.WithContext(ctx)

	req.Header = log.InjectHeader(ctx, req.Header)
	req.Header.Set("Referer", Referer)
	resp, err := HttpClient.DoWithAppSpan(req, fmt.Sprintf("%s:%s", req.Method, req.URL.Path))
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"err": err,
			"url": url,
		}).Error("API请求错误：" + url)
		code = util.ErrAPI
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"err": err,
			"url": url,
		}).Error("API请求读取response错误：" + url)
		code = util.ErrAPI
		return
	}

	return
}

func GetWithTimeOut(ctx context.Context, url string, m map[string]interface{}, timeout int64) (ret []byte, code int) {
	url = appendUrlParam(url, m)

	startTime := time.Now()
	defer func() {
		log.WithCtxFields(ctx, log.Fields{
			"url":      url,
			"timeout":  timeout,
			"duration": time.Since(startTime).Seconds(),
			"resp":     string(ret),
		}).Debug("API请求：" + url)
	}()

	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancelFunc()

	req, _ := http.NewRequestWithContext(ctxWithTimeout, "GET", url, nil)

	req.Header = log.InjectHeader(ctx, req.Header)
	req.Header.Set("Referer", Referer)
	resp, err := HttpClient.DoWithAppSpan(req, fmt.Sprintf("%s:%s", req.Method, req.URL.Path))

	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"url":     url,
			"timeout": timeout,
			"err":     err,
		}).Error("API请求错误：" + url)
		code = util.ErrAPI
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"url":     url,
			"timeout": timeout,
			"err":     err,
		}).Error("API请求读取response错误：" + url)
		code = util.ErrAPI
		return
	}

	return
}

// Post Http Post Json；
func Post(ctx context.Context, url string, m map[string]interface{}) (ret []byte, code int) {
	return PostNew(ctx, url, m)
}

// / 带超时响应请求post
func PostWithTimeOut(ctx context.Context, url string, m map[string]interface{}, timeout int64) (ret []byte, err error) {
	body, err := json.Marshal(m)
	ret = []byte{}
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"url":     url,
			"params":  m,
			"timeout": timeout,
		}).Error("API post marshal error：" + url)
		return nil, err
	}

	startTime := time.Now()
	defer func() {
		log.WithCtxFields(ctx, log.Fields{
			"url":      url,
			"params":   string(body),
			"timeout":  timeout,
			"duration": time.Since(startTime).Seconds(),
			"resp":     string(ret),
		}).Debug("API请求：" + url)
	}()

	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)

	defer cancelFunc()
	req, _ := http.NewRequestWithContext(ctxWithTimeout, "POST", url, bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header = log.InjectHeader(ctx, req.Header)
	req.Header.Set("Referer", Referer)
	resp, err := HttpClient.DoWithAppSpan(req, fmt.Sprintf("%s:%s", req.Method, req.URL.Path))

	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"url":     url,
			"params":  string(body),
			"timeout": timeout,
			"err":     err,
		}).Error("API请求错误：" + url)
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
				log.ErrCtx(ctx, fmt.Sprintf("error value from ReadAll was %T; expected some net.Error", err))
			} else if !ne.Timeout() {
				log.ErrCtx(ctx, "net.Error.Timeout = false; want true")
			}
			if got := ne.Error(); !strings.Contains(got, "Client.Timeout exceeded") {
				log.ErrCtx(ctx, fmt.Sprintf("error string = %q; missing timeout substring", got))
			}
			log.WithCtxFields(ctx, log.Fields{
				"url":     url,
				"params":  string(body),
				"timeout": timeout,
				"err":     err,
			}).Error("api read response error")
		}
	// 之前使用 <-time.After() 会导致内存暴涨
	case <-after.C:
		log.WithCtxFields(ctx, log.Fields{
			"url":     url,
			"params":  string(body),
			"timeout": timeout,
		}).Error("api read response timeout")
	}

	return
}

// PostNew Http Post Json；
func PostNew(ctx context.Context, url string, rq interface{}) (ret []byte, code int) {
	qb, err := json.Marshal(rq)
	if err != nil {
		log.Err(err.Error())
		code = util.ErrAPI
		return
	}
	ret, code = post(ctx, url, "application/json;", qb)
	return

}

func post(ctx context.Context, url string, contentType string, body []byte) (ret []byte, code int) {

	startTime := time.Now()
	defer func() {
		log.WithCtxFields(ctx, log.Fields{
			"url":      url,
			"params":   string(body),
			"duration": time.Since(startTime).Seconds(),
			"resp":     string(ret),
		}).Debug("API请求：" + url)
	}()

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header = log.InjectHeader(ctx, req.Header)
	req.Header.Set("Referer", Referer)
	resp, err := HttpClient.DoWithAppSpan(req, fmt.Sprintf("%s:%s", req.Method, req.URL.Path))

	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"url":    url,
			"params": string(body),
			"err":    err,
		}).Error("API请求错误：" + url)
		code = util.ErrAPI
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"url":    url,
			"params": string(body),
			"err":    err,
		}).Error("api read response error")
		code = util.ErrAPI
		return
	}

	return
}

// PostFormU "application/x-www-form-urlencoded"
func PostFormU(ctx context.Context, url string, m map[string]interface{}) (ret []byte, code int) {

	body := parseQuery(m)

	ret, code = post(ctx, url, "application/x-www-form-urlencoded", []byte(body))
	return

}

func reqId() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
