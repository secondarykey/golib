//OAuth access library
package oauth

import (
	"appengine"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"github.com/secondarykey/golib/appengine/web"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/*
 * OAuth認証用のType
 */
type OAuth1 struct {
	ConsumerKey    string
	ConsumerSecret string

	RequestTokenUrl   string
	AuthroizeTokenUrl string
	AccessTokenUrl    string

	RequestToken *TokenSet
	AuthroizeUrl string

	AccessToken *TokenSet

	context *appengine.Context

	authParams *web.Parameter
}

/*
 * OAuth用の定数
 */
const (
	OAUTH_VERSION    = "1.0"
	SIGNATURE_METHOD = "HMAC-SHA1"

	CALLBACK_PARAM         = "oauth_callback"
	CONSUMER_KEY_PARAM     = "oauth_consumer_key"
	NONCE_PARAM            = "oauth_nonce"
	SESSION_HANDLE_PARAM   = "oauth_session_handle"
	SIGNATURE_METHOD_PARAM = "oauth_signature_method"
	SIGNATURE_PARAM        = "oauth_signature"
	TIMESTAMP_PARAM        = "oauth_timestamp"
	TOKEN_PARAM            = "oauth_token"
	TOKEN_SECRET_PARAM     = "oauth_token_secret"
	VERIFIER_PARAM         = "oauth_verifier"
	VERSION_PARAM          = "oauth_version"
)

type TokenSet struct {
	Token  string
	Secret string
}

func NewOAuth1(key, secret, requestTokenUrl, authroizeTokenUrl, accessTokenUrl string) *OAuth1 {
	return &OAuth1{
		ConsumerKey:       key,
		ConsumerSecret:    secret,
		RequestTokenUrl:   requestTokenUrl,
		AuthroizeTokenUrl: authroizeTokenUrl,
		AccessTokenUrl:    accessTokenUrl,
	}
}

func (oauth *OAuth1) SetContext(c *appengine.Context) {
	oauth.context = c
}

/*
 * リクエストトークンの取得を行う
 *
 */
func (self *OAuth1) GetRequestToken(callback string) error {

	wb := web.NewWeb()
	//OAuth用の基本的なパラメータを作成
	self.addBaseParams()
	//CallbackURLを設定
	self.addParam(CALLBACK_PARAM, callback)

	//キーを作成
	key := escape(self.ConsumerSecret) + "&" + escape("")
	//リクエスト用の基本文字列を作成
	base := self.requestString("GET", self.RequestTokenUrl, self.authParams)
	sign := self.sign(base, key)

	self.addParam(SIGNATURE_PARAM, sign)

	data, err := self.getBody(wb, self.RequestTokenUrl)
	if err != nil {
		return err
	}

	token := data[TOKEN_PARAM]
	secret := data[TOKEN_SECRET_PARAM]

	self.RequestToken = &TokenSet{token[0], secret[0]}

	//認証URLを作成
	self.AuthroizeUrl = self.AuthroizeTokenUrl + "?" + TOKEN_PARAM + "=" + escape(token[0])
	return nil
}

/*
 * 処理メソッド、URL、引数から連結した文字列を作成
 */
func (self *OAuth1) requestString(method string, url string, args *web.Parameter) string {
	ret := method + "&" + escape(url)
	esp := "&"
	for _, key := range args.Keys() {
		ret += esp
		ret += escape(key + "=" + args.Get(key))
		esp = escape("&")
	}
	return ret
}

/*
 * HMAC-SHA1暗号を行う
 */
func (self *OAuth1) sign(message, key string) string {
	hashfun := hmac.New(sha1.New, []byte(key))
	hashfun.Write([]byte(message))
	signature := hashfun.Sum(nil)
	base64sig := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(base64sig, signature)
	return string(base64sig)
}

/*
 * OAuth用の基本的なパラメータを作成
 */
func (self *OAuth1) addBaseParams() {

	self.authParams = web.NewParameter()
	//現在時刻を設定
	clock := time.Now()
	ts := clock.Unix()
	nonce := rand.New(rand.NewSource(clock.UnixNano())).Int63()

	self.addParam(VERSION_PARAM, OAUTH_VERSION)
	self.addParam(SIGNATURE_METHOD_PARAM, SIGNATURE_METHOD)
	self.addParam(TIMESTAMP_PARAM, strconv.FormatInt(ts, 10))
	self.addParam(NONCE_PARAM, strconv.FormatInt(nonce, 10))
	self.addParam(CONSUMER_KEY_PARAM, self.ConsumerKey)

	return
}

/*
 * URLからデータを取得してパラメータを解析してmapを返す
 * リクエストトークン、アクセストークン取得用
 */
func (self *OAuth1) getBody(wb *web.Web, accessUrl string) (map[string][]string, error) {

	//認証ヘッダを設定
	wb.AddHeader("Authorization", self.getOAuthHeader())
	//指定された
	resp, err := wb.Get(accessUrl)
	if err != nil {
		return nil, err
	}

	//すべて読み込む
	bodyByte, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	//レスポンスを取得してコードを取得
	body := string(bodyByte)
	//文字列を解析
	parts, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}
	return parts, nil
}

/*
 * アクセストークンを作成
 */
func (self *OAuth1) GetAccessToken(code string) error {

	wb := web.NewWeb()
	self.addBaseParams()

	//認可コードを設定
	self.addParam(VERIFIER_PARAM, code)
	//リクエストトークンをパラメータで設定
	self.addParam(TOKEN_PARAM, self.RequestToken.Token)

	//Secretでキーを作成
	key := escape(self.ConsumerSecret) + "&" + escape(self.RequestToken.Secret)
	//
	base := self.requestString("GET", self.AccessTokenUrl, self.authParams)
	//シグネーチャーを設定
	sign := self.sign(base, key)

	self.addParam(SIGNATURE_PARAM, sign)
	//ボディーの設定
	data, err := self.getBody(wb, self.AccessTokenUrl)
	if err != nil {
		return err
	}

	//Token,Secretを取得
	token := data[TOKEN_PARAM]
	secret := data[TOKEN_SECRET_PARAM]

	//アクセストークンを作成
	self.AccessToken = &TokenSet{token[0], secret[0]}
	return nil
}

/*
 * OAuthでのGET
 */
func (self *OAuth1) Get(url string, args map[string]string) (*http.Response, error) {
	wb := self.createOAuthWeb("GET", url, args)
	return wb.Get(url)
}

/*
 * OAuthでのPOST
 */
func (self *OAuth1) Post(url string, args map[string]string) (*http.Response, error) {
	wb := self.createOAuthWeb("POST", url, args)
	return wb.Post(url)
}

func (self *OAuth1) Put(url string, data []byte) (*http.Response, error) {
	wb := self.createOAuthWeb("PUT", url, nil)
	return wb.Put(url, data)
}

/*
 * OAuthでのアクセス用のWebを取得する
 */
func (self *OAuth1) createOAuthWeb(method string, url string, args map[string]string) *web.Web {

	self.addBaseParams()
	self.addParam(TOKEN_PARAM, self.AccessToken.Token)

	//認証用のパラメータのコピーを取得
	param := self.authParams.Copy()
	//Webアクセスを生成
	wb := web.NewWeb()
	if args != nil {
		for key, value := range args {
			//Web側の引数に追加
			wb.AddParam(key, value)
			//認証用のパラメータに追加
			param.Add(key, value)
		}
	}

	wb.SetContext(self.context)

	//キーを作成
	key := escape(self.ConsumerSecret) + "&" + escape(self.AccessToken.Secret)
	base := self.requestString(method, url, param)
	sign := self.sign(base, key)

	//シグネーチャーパラムを設定
	self.addParam(SIGNATURE_PARAM, sign)
	//認証ヘッダを作成する
	wb.AddHeader("Authorization", self.getOAuthHeader())
	return wb
}

/*
 * パラメータの追加
 */
func (self *OAuth1) addParam(key, value string) {
	self.authParams.Add(key, value)
}

/*
 * OAuthヘッダの文字列を作成
 */
func (self *OAuth1) getOAuthHeader() string {
	hdr := "OAuth "
	for pos, key := range self.authParams.Keys() {
		if pos > 0 {
			hdr += ","
		}
		hdr += key + "=\"" + self.authParams.Get(key) + "\""
	}
	return hdr
}

/*
 * URLエスケープ
 */
func escape(param string) string {
	return url.QueryEscape(param)
}
