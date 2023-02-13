// Copyright 2019 The OpenZipkin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zipkin

import (
	"errors"
	"net/http"

	zipkin "github.com/openzipkin/zipkin-go"
)

// ErrValidTracerRequired error
var ErrValidTracerRequired = errors.New("valid tracer required")

// Client holds a Zipkin instrumented HTTP Client.
type Client struct {
	*http.Client
	tracer           *zipkin.Tracer
	httpTrace        bool
	defaultTags      map[string]string
	transportOptions []TransportOption
}

// ClientOption allows optional configuration of Client.
type ClientOption func(*Client)

// WithClient allows one to add a custom configured http.Client to use.
func WithClient(client *http.Client) ClientOption {
	return func(c *Client) {
		if client == nil {
			client = &http.Client{}
		}
		c.Client = client
	}
}

// ClientTrace allows one to enable Go's net/http/httptrace.
func ClientTrace(enabled bool) ClientOption {
	return func(c *Client) {
		c.httpTrace = enabled
	}
}

// ClientTags adds default Tags to inject into client application spans.
func ClientTags(tags map[string]string) ClientOption {
	return func(c *Client) {
		c.defaultTags = tags
	}
}

// TransportOptions passes optional Transport configuration to the internal
// transport used by Client.
func TransportOptions(options ...TransportOption) ClientOption {
	return func(c *Client) {
		c.transportOptions = options
	}
}

// NewClient returns an HTTP Client adding Zipkin instrumentation around an
// embedded standard Go http.Client.
func NewClient(tracer *zipkin.Tracer, options ...ClientOption) (*Client, error) {
	if tracer == nil {
		return nil, ErrValidTracerRequired
	}

	c := &Client{tracer: tracer, Client: &http.Client{}}
	for _, option := range options {
		option(c)
	}

	c.transportOptions = append(
		c.transportOptions,
		// the following Client settings override provided transport settings.
		RoundTripper(c.Client.Transport),
		TransportTrace(c.httpTrace),
	)
	transport, err := NewTransport(tracer, c.transportOptions...)
	if err != nil {
		return nil, err
	}
	c.Client.Transport = transport

	return c, nil
}

// DoWithAppSpan wraps http.Client's Do with tracing using an application span.
func (c *Client) DoWithAppSpan(req *http.Request, name string) (res *http.Response, err error) {
	res, err = c.Client.Do(req)
	if err != nil {
		return
	}

	return
}
