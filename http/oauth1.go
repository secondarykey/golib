package http

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	original "net/http"
	"net/url"
	"strconv"
	"time"
)

type OAuth1 struct {
	Credential *Credential

	requestTokenUrl string
	authorizeUrl    string
	accessTokenUrl  string

	requestToken  string
	requestSecret string

	authParams *parameter
}

type Credential struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

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

func NewOAuth1(c *Credential, request, authorize, access string) *OAuth1 {
	oa := &OAuth1{
		Credential:      c,
		requestTokenUrl: request,
		authorizeUrl:    authorize,
		accessTokenUrl:  access,
	}
	return oa
}

func (o *OAuth1) GetRequestToken(callback string) error {

	wb := NewClient()
	o.addBaseParams()
	o.addParam(CALLBACK_PARAM, callback)

	key := escape(o.Credential.ConsumerSecret) + "&" + escape("")
	base := o.requestString("GET", o.requestTokenUrl, o.authParams)
	sign := o.sign(base, key)

	o.addParam(SIGNATURE_PARAM, sign)

	data, err := o.getBody(wb, o.requestTokenUrl)
	if err != nil {
		return fmt.Errorf("getBody Error:%s", err)
	}

	token := data[TOKEN_PARAM]
	secret := data[TOKEN_SECRET_PARAM]

	o.requestToken = token[0]
	o.requestSecret = secret[0]

	return nil
}

func (o *OAuth1) GetAuthorizeURL() string {
	return o.authorizeUrl + "?" + TOKEN_PARAM + "=" + escape(o.requestToken)
}

func (o *OAuth1) requestString(method string, url string, args *parameter) string {
	ret := method + "&" + escape(url)
	esp := "&"
	for _, key := range args.keys() {
		ret += esp
		ret += escape(key + "=" + args.get(key))
		esp = escape("&")
	}
	return ret
}

func (o *OAuth1) sign(message, key string) string {
	hashfun := hmac.New(sha1.New, []byte(key))
	hashfun.Write([]byte(message))
	signature := hashfun.Sum(nil)
	base64sig := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(base64sig, signature)
	return string(base64sig)
}

func (o *OAuth1) addBaseParams() {

	o.authParams = NewParameter()

	clock := time.Now()
	ts := clock.Unix()
	nonce := rand.New(rand.NewSource(clock.UnixNano())).Int63()

	o.addParam(VERSION_PARAM, OAUTH_VERSION)
	o.addParam(SIGNATURE_METHOD_PARAM, SIGNATURE_METHOD)
	o.addParam(TIMESTAMP_PARAM, strconv.FormatInt(ts, 10))
	o.addParam(NONCE_PARAM, strconv.FormatInt(nonce, 10))
	o.addParam(CONSUMER_KEY_PARAM, o.Credential.ConsumerKey)

	return
}

func (o *OAuth1) getBody(wb *Client, accessUrl string) (map[string][]string, error) {

	wb.header.Add("Authorization", o.getOAuthHeader())

	resp, err := wb.Get(accessUrl)
	if err != nil {
		return nil, fmt.Errorf("Web Get Error:%s", err)
	}
	defer resp.Body.Close()

	bodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Response Read Error:%s", err)
	}

	body := string(bodyByte)
	parts, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}
	return parts, nil
}

func (o *OAuth1) GetAccessToken(code string) error {

	wb := NewClient()
	o.addBaseParams()

	o.addParam(VERIFIER_PARAM, code)
	o.addParam(TOKEN_PARAM, o.requestToken)

	key := escape(o.Credential.ConsumerSecret) + "&" + escape(o.requestSecret)
	base := o.requestString("GET", o.accessTokenUrl, o.authParams)
	sign := o.sign(base, key)

	o.addParam(SIGNATURE_PARAM, sign)

	data, err := o.getBody(wb, o.accessTokenUrl)
	if err != nil {
		return fmt.Errorf("getBody Error:%s", err)
	}

	token := data[TOKEN_PARAM]
	secret := data[TOKEN_SECRET_PARAM]

	o.Credential.AccessToken = token[0]
	o.Credential.AccessSecret = secret[0]

	return nil
}

func (o *OAuth1) Get(url string, args map[string]string) (*original.Response, error) {
	wb, err := o.createOAuthParameter("GET", url, args)
	if err != nil {
		return nil, fmt.Errorf("Create OAuth Paramerter Error:%s", err)
	}
	return wb.Get(url)
}

func (o *OAuth1) Post(url string, args map[string]string) (*original.Response, error) {
	wb, err := o.createOAuthParameter("POST", url, args)
	if err != nil {
		return nil, fmt.Errorf("Create OAuth Paramerter Error:%s", err)
	}
	return wb.Post(url)
}

func (o *OAuth1) createOAuthParameter(method string, url string, args map[string]string) (*Client, error) {

	if o.Credential.ConsumerKey == "" || o.Credential.ConsumerSecret == "" {
		return nil, fmt.Errorf("CustomerKey,Secret is empty!")
	}

	if o.Credential.AccessToken == "" || o.Credential.AccessSecret == "" {
		return nil, fmt.Errorf("AccessToken,Secret is empty!")
	}

	o.addBaseParams()
	o.addParam(TOKEN_PARAM, o.Credential.AccessToken)

	param := o.authParams.copy()
	wb := NewClient()
	if args != nil {
		for key, value := range args {
			wb.AddParam(key, value)
			param.add(key, value)
		}
	}

	key := escape(o.Credential.ConsumerSecret) + "&" + escape(o.Credential.AccessSecret)
	base := o.requestString(method, url, param)
	sign := o.sign(base, key)

	o.addParam(SIGNATURE_PARAM, sign)

	wb.header.Add("Authorization", o.getOAuthHeader())

	return wb, nil
}

func (o *OAuth1) addParam(key, value string) {
	o.authParams.add(key, value)
}

func (o *OAuth1) getOAuthHeader() string {
	hdr := "OAuth "
	for pos, key := range o.authParams.keys() {
		if pos > 0 {
			hdr += ","
		}
		hdr += key + "=\"" + o.authParams.get(key) + "\""
	}
	return hdr
}
