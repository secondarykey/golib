package web

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestWeb(t *testing.T) {

	Convey("escape() Test", t, func() {
		So(escape("日本語"), ShouldEqual, "%E6%97%A5%E6%9C%AC%E8%AA%9E")
		So(escape("="), ShouldEqual, "%3D")
	})

	Convey("Web Test", t, func() {

		Convey("Web生成", func() {
			wb := NewWeb()
			So(wb.contentType, ShouldEqual, "")
			So(wb.header, ShouldNotBeNil)
			So(wb.params, ShouldNotBeNil)
		})

		Convey("存在するWebにアクセス", func() {
			wb := NewWeb()
			resp, err := wb.Get("http://www.yahoo.co.jp")
			So(resp.StatusCode, ShouldEqual, 200)
			So(err, ShouldBeNil)
		})

		Convey("存在しないWebにアクセス", func() {
			wb := NewWeb()
			resp, err := wb.Get("http://www.test")
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
	})

	Convey("HttpError Test", t, func() {
		Convey("表示の確認", func() {
			err := HttpError{"正常", 200}
			So(err.Error(), ShouldEqual, "200:\n正常")
		})
	})

	Convey("Parameter Test", t, func() {

		Convey("生成のテスト", func() {
			params := NewParameter()
			So(params.param, ShouldNotBeNil)
			So(len(params.param), ShouldEqual, 0)
			So(len(params.order), ShouldEqual, 0)
		})

		Convey("追加と取得", func() {
			params := NewParameter()
			Convey("通常動作", func() {
				params.Add("key", "value")
				So(len(params.param), ShouldEqual, 1)
				So(len(params.order), ShouldEqual, 1)
				So(params.Get("key"), ShouldEqual, "value")

				params.Add("key2", "value2")
				So(len(params.param), ShouldEqual, 2)
				So(len(params.order), ShouldEqual, 2)
				So(params.Get("key2"), ShouldEqual, "value2")
			})

			Convey("同一のキーは追加されない", func() {
				params.Add("key2", "value")
				So(len(params.param), ShouldEqual, 2)
				So(len(params.order), ShouldEqual, 2)
				So(params.Get("key2"), ShouldEqual, "value2")
			})

			Convey("日本語はエスケープされる", func() {
				params.Add("key3", "日本語")
				So(params.Get("key3"), ShouldEqual, "%E6%97%A5%E6%9C%AC%E8%AA%9E")
			})

			Convey("addUnEscapeでエスケープされない", func() {
				params.addUnEscape("key4", "日本語")
				So(params.Get("key4"), ShouldEqual, "日本語")
			})
		})

		Convey("keys()のソートテスト", func() {
			params := NewParameter()
			params.Add("ccc", "Value")
			params.Add("aaa", "Value")
			params.Add("BBB", "Value")
			params.Add("AAA", "Value")
			p := params.Keys()
			keys := []string{"AAA", "BBB", "aaa", "ccc"}
			So(p, ShouldResemble, keys)
		})

		Convey("copyのテスト", func() {
			params := NewParameter()
			params.Add("AAA", "Value")
			params.Add("BBB", "Value")
			params.Add("aaa", "Value")
			params.Add("ccc", "Value")

			clone := params.Copy()
			//元に追加してみる
			params.Add("ddd", "Value")
			Convey("コピー側は元の通り", func() {
				So(len(clone.param), ShouldEqual, 4)
				So(len(clone.order), ShouldEqual, 4)
				p := clone.Keys()
				keys := []string{"AAA", "BBB", "aaa", "ccc"}
				So(p, ShouldResemble, keys)
			})
			Convey("追加した方は変更される", func() {
				So(len(params.param), ShouldEqual, 5)
				So(len(params.order), ShouldEqual, 5)
			})
		})

	})

}
