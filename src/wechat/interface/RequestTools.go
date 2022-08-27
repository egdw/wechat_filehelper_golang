package _interface

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func GET(wechatFileHelper *WechatFileHelper, url string) (string, []byte) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header = http.Header{
		"mmweb_appid": {"wx_webfilehelper"},
	}
	response, err := wechatFileHelper.Client.Do(req)
	defer response.Body.Close()
	if err != nil {
		println(err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		println(err.Error())
	}
	return string(body), body

}

func POST(wechatFileHelper *WechatFileHelper, url string, data url.Values) string {
	response, err := wechatFileHelper.Client.PostForm(url, data)
	defer response.Body.Close()
	if err != nil {
		println(err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		println(err.Error())
	}
	return string(body)
}

func POST_JSON_BODY(wechatFileHelper *WechatFileHelper, url string, json string) string {

	req, err := http.NewRequest("POST", url, strings.NewReader(json))
	if err != nil {
		panic(err)
	}
	req.Header = http.Header{
		"mmweb_appid":  {"wx_webfilehelper"},
		"Content-Type": {"application/json;charset=UTF-8"},
		"Accept":       {"application/json, text/plain, */*"},
	}
	response, err := wechatFileHelper.Client.Do(req)
	defer response.Body.Close()
	if err != nil {
		println(err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		println(err.Error())
	}
	return string(body)
}

//post的同时获取header
func POST_JSON_BODY_Header(wechatFileHelper *WechatFileHelper, url string, json string) (string, string) {

	req, err := http.NewRequest("POST", url, strings.NewReader(json))
	if err != nil {
		panic(err)
	}
	req.Header = http.Header{
		"mmweb_appid":  {"wx_webfilehelper"},
		"Content-Type": {"application/json;charset=UTF-8"},
		"Accept":       {"application/json, text/plain, */*"},
	}
	response, err := wechatFileHelper.Client.Do(req)
	defer response.Body.Close()
	if err != nil {
		println(err.Error())
	}
	var webwx_data_ticket = ""
	for _, cookie := range response.Cookies() {
		if cookie.Name == "webwx_data_ticket" {
			webwx_data_ticket = cookie.Value
		}
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		println(err.Error())
	}
	return string(body), webwx_data_ticket
}
