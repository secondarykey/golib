package http

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type Client struct {
	contentType string
	header      http.Header
	params      *parameter
}

type parameter struct {
	param map[string]string
	order []string
}

func escape(param string) string {
	return url.QueryEscape(param)
}

func NewClient() *Client {
	return &Client{
		params:      NewParameter(),
		header:      http.Header{},
		contentType: "",
	}
}

func (w *Client) AddParam(key, value string) {
	w.params.add(key, value)
}

func NewParameter() *parameter {
	return &parameter{
		param: make(map[string]string),
		order: make([]string, 0),
	}
}

func (p *parameter) add(key, value string) {
	p.addUnEscape(key, escape(value))
}

func (p *parameter) addUnEscape(key, value string) {
	if _, flag := p.param[key]; !flag {
		p.param[key] = value
		p.order = append(p.order, key)
	}
}

func (p *parameter) get(key string) string {
	return p.param[key]
}

func (p *parameter) copy() *parameter {
	clone := NewParameter()
	for _, key := range p.keys() {
		clone.addUnEscape(key, p.get(key))
	}
	return clone
}

func (p *parameter) keys() []string {
	sort.Strings(p.order)
	return p.order
}

func (w *Client) getQuery() string {
	params := w.params.keys()
	ret := ""
	sep := ""
	for _, key := range params {
		value := w.params.get(key)
		ret += sep + key + "=" + value
		sep = "&"
	}
	return ret
}

func (w *Client) Get(url string) (*http.Response, error) {
	q := w.getQuery()
	if q != "" {
		q = "?" + q
	}
	return w.execute("GET", url+q, "")
}

func (w *Client) Post(url string) (*http.Response, error) {
	w.contentType = "application/x-www-form-urlencoded"
	return w.execute("POST", url, w.getQuery())
}

func (w *Client) execute(method string, url string, body string) (*http.Response, error) {

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Create Request Error:%s", err)
	}

	req.Header = w.header
	if w.contentType != "" {
		req.Header.Set("Content-Type", w.contentType)
	}
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		defer resp.Body.Close()
		return nil, fmt.Errorf("StatusError[%d]%s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}
