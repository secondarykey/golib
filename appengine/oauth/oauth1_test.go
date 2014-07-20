package oauth

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestOAuth(t *testing.T) {

	var tokenSet TokenSet
	tokenSet.Token = "aaa"
	tokenSet.Token = "bbb"

	//テストの方法を考える
	oa := NewOAuth1(
		tokenSet.Token,
		tokenSet.Secret,
		"https://api.twitter.com/oauth/request_token",
		"https://api.twitter.com/oauth/authorize",
		"https://api.twitter.com/oauth/access_token")
	oa.GetRequestToken("oob")

	Convey("OAuth Test", t, func() {
		Convey("RequestTokenの取得", func() {
		})
	})

}
