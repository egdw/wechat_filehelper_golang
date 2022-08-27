package main

import (
	_ "net/http"
	WechatFileHelperTools "wechat_filehelper_golang/src/wechat/interface"
)

func main() {
	wechatFileHelper := WechatFileHelperTools.Init()
	wechatFileHelper.Login()
	wechatFileHelper.SyncCheck()
	println(&wechatFileHelper)
}
