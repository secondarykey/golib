package web

import (
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

/*
 * Webアクセス用のタイプ
 * contentType : コンテンツタイプ
 * header : リクエスト時のヘッダを指定
 * params : 引数(AddParamで設定)
 */
type Web struct {
	contentType string
	header      http.Header
	params      *Parameter
}

/*
 * パラメータ用のタイプ
 */
type Parameter struct {
	param map[string]string
	order []string
}

/*
 * URLエスケープ
 */
func escape(param string) string {
	return url.QueryEscape(param)
}

/*
 * HTTP エラーコード
 */
type HttpError struct {
	status     string
	statusCode int
}

/*
 * エラー実装
 */
func (self HttpError) Error() string {
	return strconv.Itoa(self.statusCode) + ":\n" +
		self.status
}

/*
 * Webインスタンスの生成
 */
func NewWeb() *Web {
	return &Web{
		params:      NewParameter(),
		header:      http.Header{},
		contentType: "",
	}
}

/*
 * パラメータの追加
 */
func (self *Web) AddParam(key, value string) {
	self.params.Add(key, value)
}

func (self *Web) AddHeader(key, value string) {
	self.header.Add(key, value)
}

/*
 * パラメータ用のType
 * paramは保存場所
 * orderは引数のソート用に使用する
 */
func NewParameter() *Parameter {
	return &Parameter{
		param: make(map[string]string),
		order: make([]string, 0),
	}
}

/*
 * 引数の追加を行う
 */
func (self *Parameter) Add(key, value string) {
	self.addUnEscape(key, escape(value))
}

/*
 * addのエスケープをしない場合の呼び出し
 * 通常copyの時しか使わない
 */
func (self *Parameter) addUnEscape(key, value string) {
	if _, flag := self.param[key]; !flag {
		self.param[key] = value
		self.order = append(self.order, key)
	}
}

/*
 * 値の取得
 */
func (self *Parameter) Get(key string) string {
	return self.param[key]
}

/*
 * 新規にparameterを生成して、自身のコピーをする
 */
func (self *Parameter) Copy() *Parameter {
	clone := NewParameter()
	for _, key := range self.Keys() {
		clone.addUnEscape(key, self.Get(key))
	}
	return clone
}

/*
 * キーの取り出しsort.Strings()でソートしてから返す
 */
func (self *Parameter) Keys() []string {
	sort.Strings(self.order)
	return self.order
}

/*
 * リクエストパラメータを設定
 */
func (self *Web) getQuery() string {
	params := self.params.Keys()
	ret := ""
	sep := ""
	for _, key := range params {
		value := self.params.Get(key)
		ret += sep + key + "=" + value
		sep = "&"
	}
	return ret
}

/*
 * Webページの取得を行う(GET)
 */
func (self *Web) Get(url string) (*http.Response, error) {
	q := self.getQuery()
	if q != "" {
		q = "?" + q
	}
	return self.execute("GET", url+q, "")
}

/*
 * Webページの取得を行う(POST)
 */
func (self *Web) Post(url string) (*http.Response, error) {
	self.contentType = "application/x-www-form-urlencoded"
	return self.execute("POST", url, self.getQuery())
}

/*
 * methodに応じた処理を行う
 * 現状サポートはGET、POSTのみ
 */
func (self *Web) execute(method string, url string, body string) (*http.Response, error) {

	//リクエストの生成
	req, reqErr := http.NewRequest(method, url, strings.NewReader(body))
	if reqErr != nil {
		return nil, reqErr
	}
	//ヘッダの設定
	req.Header = self.header

	//コンテンツタイプがある場合は設定
	if self.contentType != "" {
		req.Header.Set("Content-Type", self.contentType)
	}
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	client := &http.Client{}
	//リクエストによる処理を行う
	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, doErr
	}

	//ステータスコードに応じてエラー処理をする
	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		defer resp.Body.Close()
		return nil, HttpError{
			status:     resp.Status,
			statusCode: resp.StatusCode,
		}
	}
	return resp, nil
}
