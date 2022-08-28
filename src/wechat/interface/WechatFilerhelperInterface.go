package _interface

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Han-Ya-Jun/qrcode2console"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 二维码ID提取正则
var qrCodeRegex = regexp.MustCompile("\\S{10}==")

// 查找URL正则表达式
var findUrlRegex = regexp.MustCompile("(https?://[a-zA-Z0-9\\.\\?/%-_]*)")

var retCodeRegex = regexp.MustCompile("retcode:\"(.*?)\"")

var selectorRegex = regexp.MustCompile("selector:\"(.*?)\"")

type WechatFileHelper struct {
	zf         bool
	Client     http.Client
	SKey       string
	WxSid      string
	WxUin      int
	PassTicket string
	SyncKey    struct {
		Count int `json:"Count"`
		List  []struct {
			Key int `json:"Key"`
			Val int `json:"Val"`
		} `json:"List"`
	} `json:"SyncKey"`
	User struct {
		UserName   string `json:"UserName"`
		NickName   string `json:"NickName"`
		HeadImgUrl string `json:"HeadImgUrl"`
	} `json:"User"`
	xsyncUri string
	username string
}

func (wechatFileHelper *WechatFileHelper) Login() {
	GET(wechatFileHelper, "https://filehelper.weixin.qq.com/") //初始化cookie数据
	//println(index)
	var scaning bool = true
	for scaning {
		queryTime := 0
		qrCode, _ := GET(wechatFileHelper, "https://login.wx.qq.com/jslogin?appid=wx_webfilehelper&redirect_uri=https%253A%252F%252Ffilehelper.weixin.qq.com%252Fcgi-bin%252Fmmwebwx-bin%252Fwebwxnewloginpage&fun=new&lang=zh_CN")
		if strings.Index(qrCode, "200") >= 0 {
			//二维码ok，获取uuid
			qrcodeId := qrCodeRegex.FindString(qrCode)
			//println(qrcodeId)
			_, qrImg := GET(wechatFileHelper, "https://login.weixin.qq.com/qrcode/"+qrcodeId)

			err := ioutil.WriteFile("./qrcode.png", qrImg, 0777)
			if err != nil {
				panic("二维码写入失败")
			}
			//针对没有页面的进行输出，有页面的直接打开qrcode.png
			qrcode2console.NewQRCode2ConsoleWithPath("./qrcode.png").Output()
			redirectUri := ""
			for queryTime <= 25 { //二维码有效期20秒钟
				scanStatus, _ := GET(wechatFileHelper, "https://login.wx.qq.com/cgi-bin/mmwebwx-bin/login?loginicon=true&uuid="+qrcodeId+"&tip=1&appid=wx_webfilehelper")
				if strings.Index(scanStatus, "408") >= 0 {
					//还没有登录成功等待2秒
					time.Sleep(5 * time.Second)
				} else if strings.Index(scanStatus, "201") >= 0 {
					//说明用户扫描了。
					println("检测到用户扫描二维码！请在手机上确认登录！")
					innerqueryTime := 0
					for true {
						scanStatus, _ := GET(wechatFileHelper, "https://login.wx.qq.com/cgi-bin/mmwebwx-bin/login?loginicon=true&uuid="+qrcodeId+"&tip=1&appid=wx_webfilehelper")
						if strings.Index(scanStatus, "window.code=200") >= 0 {
							//说明手机确认登录成功
							println(scanStatus)
							redirectUri = findUrlRegex.FindString(scanStatus)
							if strings.Index(redirectUri, "szfilehelper") < 0 { //如果不是新版的，后面要加这几个参数
								redirectUri = redirectUri + "&fun=new&version=v2"
							}
							scaning = false
							break
						} else {
							time.Sleep(1 * time.Second)
							if innerqueryTime > 15 {
								break
							}
							innerqueryTime += 1
						}
					}
					if !scaning {
						break
					}
					if queryTime >= 10 {
						println("确认登录操作超时，请重新扫描二维码！")
						break
					}
				} else {
					println(scanStatus)
				}
				queryTime += 5

			}
			if !scaning {
				//说明用户已经确认了。
				//请求init
				println(redirectUri)
				//判断是否是新版的微信文件传输助手，如果是新版的微信文件传输助手，那么在redirectUri就无法获取到数据，
				//需要从Init调用数据。
				wechatFileHelper.zf = false
				if strings.Index(redirectUri, "szfilehelper") >= 0 {
					wechatFileHelper.zf = true
					_, _, _, webwx_data_ticket, wxuin, wxsid := GET_new_header(wechatFileHelper, redirectUri)
					println("新版zfFilehelper方案")
					//是从cookie中直接提取id，而不是从返回里面提取id
					wechatFileHelper.SKey = "" //默认Skey为null
					wechatFileHelper.WxSid = wxsid
					atoi, _ := strconv.Atoi(wxuin)
					wechatFileHelper.WxUin = atoi
					wechatFileHelper.PassTicket = webwx_data_ticket
				} else {
					//获取对应的Skey Wxsid Wxuin等数据
					infoPage, _ := GET(wechatFileHelper, redirectUri)
					webwxnewloginpage := Webwxnewloginpage{}
					xml.Unmarshal([]byte(infoPage), &webwxnewloginpage)
					if webwxnewloginpage.Ret == 0 { //将相关数据填充到内部
						wechatFileHelper.SKey = webwxnewloginpage.Skey
						wechatFileHelper.WxSid = webwxnewloginpage.WxSid
						wechatFileHelper.WxUin = webwxnewloginpage.WxUin
						wechatFileHelper.PassTicket = webwxnewloginpage.PassTicket
					}
				}

				//进行init请求的数据提取
				//主要包括SyncKey的提取
				var marshal = []byte{}
				webwxInitRetBody := WebwxInitRetBody{}
				if wechatFileHelper.zf {
					marshal, _ = json.Marshal(wechatFileHelper.GetSZFileBaseRequest())
					err = json.Unmarshal([]byte(POST_JSON_BODY(wechatFileHelper, "https://szfilehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxinit?lang=zh_CN&pass_ticket="+url.QueryEscape(wechatFileHelper.PassTicket), string(marshal))), &webwxInitRetBody)
					if err != nil {
						panic(err)
					}
				} else {
					marshal, _ = json.Marshal(wechatFileHelper.GetBaseRequest())
					err = json.Unmarshal([]byte(POST_JSON_BODY(wechatFileHelper, "https://filehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxinit?lang=zh_CN&pass_ticket="+url.QueryEscape(wechatFileHelper.PassTicket), string(marshal))), &webwxInitRetBody)
					if err != nil {
						panic(err)
					}
				}
				//相关数据填充到wechatFileHelper中，提醒用户登录成功
				wechatFileHelper.User = webwxInitRetBody.User
				wechatFileHelper.SyncKey = webwxInitRetBody.SyncKey
				wechatFileHelper.username = webwxInitRetBody.User.UserName
				if wechatFileHelper.zf {
					wechatFileHelper.SKey = webwxInitRetBody.SKey
					wechatFileHelper.xsyncUri = "https://szfilehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxsync?sid=" + url.QueryEscape(wechatFileHelper.WxSid) + "&skey=" + url.QueryEscape(wechatFileHelper.SKey) + "&pass_ticket=" + url.QueryEscape(wechatFileHelper.PassTicket)
				} else {
					wechatFileHelper.xsyncUri = "https://filehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxsync?sid=" + url.QueryEscape(wechatFileHelper.WxSid) + "&skey=" + url.QueryEscape(wechatFileHelper.SKey) + "&pass_ticket=" + url.QueryEscape(wechatFileHelper.PassTicket)
				}
				println("登录成功，欢迎您" + wechatFileHelper.User.NickName)
				break
			}
		}

	}

}

// ProcessMessage 处理文件传输助手变动的消息
func (wechatFileHelper *WechatFileHelper) ProcessMessage() bool {
	syncBody := XsyncBody{
		BaseRequest: wechatFileHelper.GetBaseRequest().BaseRequest,
		SyncKey:     wechatFileHelper.SyncKey,
	}
	marshal, _ := json.Marshal(syncBody)
	body, webwxDataTicket := POST_JSON_BODY_Header(wechatFileHelper, wechatFileHelper.xsyncUri, string(marshal))
	xSyncBody := SyncBody{}
	json.Unmarshal([]byte(body), &xSyncBody)
	if xSyncBody.BaseResponse.Ret != 0 {
		//说明无法获取消息了。
		return false
	}
	if xSyncBody.AddMsgCount > 0 {
		for i := 0; i < xSyncBody.AddMsgCount; i++ {
			xSync := xSyncBody.AddMsgList[i]
			switch xSyncBody.AddMsgList[i].MsgType {
			case 1: //文本
				wechatFileHelper.SendTextMessage("收到消息：" + xSyncBody.AddMsgList[i].Content)
				break
			case 3: //img
				downLoadImgUri := ""
				if wechatFileHelper.zf {
					downLoadImgUri = fmt.Sprintf("https://szfilehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxgetmsgimg?MsgID=%s&skey=%s&mmweb_appid=wx_webfilehelper",
						xSync.MsgId, url.QueryEscape(wechatFileHelper.SKey))
				} else {
					downLoadImgUri = fmt.Sprintf("https://filehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxgetmsgimg?MsgID=%s&skey=%s&mmweb_appid=wx_webfilehelper",
						xSync.MsgId, url.QueryEscape(wechatFileHelper.SKey))
				}

				_, downloadBytes := GET(wechatFileHelper, downLoadImgUri)
				filename := time.Now().Format("20060102150405") + "_" + GetRandomString(6) + ".jpg"
				//ioutil.WriteFile("E:\\罗宾阿里云盘\\微信文件助手\\图片\\"+filename, downloadBytes, 0777)
				ioutil.WriteFile("./"+filename, downloadBytes, 0777)

				wechatFileHelper.SendTextMessage("图片" + filename + "已保存！")
				break
			case 34: //voice
				println("接收到声音：" + string(xSync.VoiceLength))
				wechatFileHelper.SendTextMessage("我收到了" + wechatFileHelper.User.NickName + "那充满慈爱的声音！")
				break
			case 43: //video
				println("接收到视频:" + string(xSync.PlayLength))
				downLoadImgUri := ""
				if wechatFileHelper.zf {
					downLoadImgUri = fmt.Sprintf("https://szfilehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxgetvideo?msgid=%s&skey=%s&mmweb_appid=wx_webfilehelper&fun=download",
						xSync.MsgId, url.QueryEscape(wechatFileHelper.SKey))
				} else {
					downLoadImgUri = fmt.Sprintf("https://filehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxgetvideo?msgid=%s&skey=%s&mmweb_appid=wx_webfilehelper&fun=download",
						xSync.MsgId, url.QueryEscape(wechatFileHelper.SKey))
				}

				_, downloadBytes := GET(wechatFileHelper, downLoadImgUri)
				filename := time.Now().Format("20060102150405") + "_" + GetRandomString(6) + ".mp4"
				//ioutil.WriteFile("E:\\罗宾阿里云盘\\微信文件助手\\视频\\"+filename, downloadBytes, 0777)
				ioutil.WriteFile("./"+filename, downloadBytes, 0777)

				wechatFileHelper.SendTextMessage("图片" + filename + "已保存！")

				break
			case 49: //file
				downLoadFileUri := ""
				if wechatFileHelper.zf {
					downLoadFileUri = fmt.Sprintf("https://file.wx2.qq.com/cgi-bin/mmwebwx-bin/webwxgetmedia?sender=%s&mediaid=%s&encryfilename=%s&fromuser=%s&pass_ticket=%s&webwx_data_ticket=%s&sid=%s&mmweb_appid=wx_webfilehelper",
						xSync.FromUserName, xSync.MediaId, xSync.EncryFileName, strconv.Itoa(wechatFileHelper.WxUin), url.QueryEscape(wechatFileHelper.PassTicket), url.QueryEscape(webwxDataTicket), url.QueryEscape(wechatFileHelper.WxSid))

				} else {
					downLoadFileUri = fmt.Sprintf("https://file.wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetmedia?sender=%s&mediaid=%s&encryfilename=%s&fromuser=%s&pass_ticket=%s&webwx_data_ticket=%s&sid=%s&mmweb_appid=wx_webfilehelper",
						xSync.FromUserName, xSync.MediaId, xSync.EncryFileName, strconv.Itoa(wechatFileHelper.WxUin), url.QueryEscape(wechatFileHelper.PassTicket), url.QueryEscape(webwxDataTicket), url.QueryEscape(wechatFileHelper.WxSid))
				}
				_, downloadBytes := GET(wechatFileHelper, downLoadFileUri)
				filename := time.Now().Format("20060102150405") + "_" + xSync.FileName
				//ioutil.WriteFile("E:\\罗宾阿里云盘\\微信文件助手\\文件\\"+filename, downloadBytes, 0777)
				ioutil.WriteFile("./"+filename, downloadBytes, 0777)
				wechatFileHelper.SendTextMessage("文件" + filename + "已保存！")

				break
			case 51: //代表用户进入文件传输助手
				wechatFileHelper.SendTextMessage("你好" + wechatFileHelper.User.NickName + ",我还活着呢~")
				break
			default:
				println("未知类型：", strconv.Itoa(xSyncBody.AddMsgList[i].MsgType))
				break
			}
		}
		//替换原始helper的数据，这样就可以自动确认接收到的消息了。
		wechatFileHelper.SyncKey = xSyncBody.SyncKey
	}
	return true
}

func (wechatFileHelper *WechatFileHelper) SyncCheck() { //同步数据
	func() {
		//循环读取同步数据
		builder := strings.Builder{}
		for true {
			builder.Reset()
			for i := 0; i < wechatFileHelper.SyncKey.Count; i++ {
				builder.WriteString(strconv.Itoa(wechatFileHelper.SyncKey.List[i].Key))
				builder.WriteByte('_')
				builder.WriteString(strconv.Itoa(wechatFileHelper.SyncKey.List[i].Val))
				if i+1 < wechatFileHelper.SyncKey.Count {
					builder.WriteByte('|')
				}
			}
			syncheckUri := ""
			if wechatFileHelper.zf {
				syncheckUri = fmt.Sprintf("https://szfilehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/synccheck?r=%s&skey=%s&sid=%s&uin=%s&deviceid=%s&synckey=%s&mmweb_appid=wx_webfilehelper",
					strconv.Itoa(int(time.Now().UnixMilli())),
					url.QueryEscape(wechatFileHelper.SKey),
					url.QueryEscape(wechatFileHelper.WxSid),
					url.QueryEscape(strconv.Itoa(wechatFileHelper.WxUin)),
					GetRandomString(15),
					url.QueryEscape(builder.String()))
			} else {
				syncheckUri = fmt.Sprintf("https://filehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/synccheck?r=%s&skey=%s&sid=%s&uin=%s&deviceid=%s&synckey=%s&mmweb_appid=wx_webfilehelper",
					strconv.Itoa(int(time.Now().UnixMilli())),
					url.QueryEscape(wechatFileHelper.SKey),
					url.QueryEscape(wechatFileHelper.WxSid),
					url.QueryEscape(strconv.Itoa(wechatFileHelper.WxUin)),
					GetRandomString(15),
					url.QueryEscape(builder.String()))
			}

			syncheckBody, _ := GET(wechatFileHelper, syncheckUri)
			if retCodeRegex.FindString(syncheckBody) == "retcode:\"0\"" {
				if strings.Index(selectorRegex.FindString(syncheckBody), "0") < 0 {
					//说明有消息，进行数据处理
					if !wechatFileHelper.ProcessMessage() {
						println("消息无法处理，登录状态消失！")
						break
					}
				}
			} else if retCodeRegex.FindString(syncheckBody) == "retcode:\"1101\"" {
				//说明登录状态消失了，需要重新登录
				println("登录状态消失了，需要重新登录！")
				break
			} else if retCodeRegex.FindString(syncheckBody) == "retcode:\"1100\"" {
				//说明有消息，进行数据处理
				if !wechatFileHelper.ProcessMessage() {
					println("消息无法处理！" + syncheckBody)
					break
				}
			} else {
				println("获取同步消息异常：" + syncheckBody)
			}
			time.Sleep(3 * time.Second) //3秒请求一次
		}
	}()
}

func Init() WechatFileHelper {
	rand.Seed(time.Now().UnixNano())
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client := http.Client{Jar: cookieJar}
	return WechatFileHelper{Client: client}
}

var letters = []rune("0123456789")

//GetRandomString ...
func GetRandomString(l int) string {
	b := make([]rune, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// SendTextMessage 文件传输助手发送文本信息
func (wechatFileHelper *WechatFileHelper) SendTextMessage(content string) {
	msgId := strconv.FormatInt(time.Now().UnixNano(), 10)
	sendTextBody := SendTextBody{
		BaseRequest: struct {
			Uin      int    `json:"Uin"`
			Sid      string `json:"Sid"`
			Skey     string `json:"Skey"`
			DeviceID string `json:"DeviceID"`
		}{Uin: wechatFileHelper.WxUin, Sid: wechatFileHelper.WxSid, Skey: wechatFileHelper.SKey, DeviceID: GetRandomString(15)},
		Msg: struct {
			ClientMsgId  string `json:"ClientMsgId"`
			FromUserName string `json:"FromUserName"`
			LocalID      string `json:"LocalID"`
			ToUserName   string `json:"ToUserName"`
			Content      string `json:"Content"`
			Type         int    `json:"Type"`
		}{ClientMsgId: msgId, FromUserName: wechatFileHelper.username, LocalID: msgId, ToUserName: "filehelper", Type: 1, Content: content},
		Scene: 0,
	}
	marshal, _ := json.Marshal(sendTextBody)
	messageSentRetBody := MessageSentRetBody{}
	if wechatFileHelper.zf {
		err := json.Unmarshal([]byte(POST_JSON_BODY(wechatFileHelper, "https://szfilehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsg?lang=zh_CN&pass_ticket="+url.QueryEscape(wechatFileHelper.PassTicket), string(marshal))), &messageSentRetBody)
		if err != nil {
			panic(err)
		}
		if messageSentRetBody.BaseResponse.Ret != 0 {
			println("发送消息失败！")
		}
	} else {
		err := json.Unmarshal([]byte(POST_JSON_BODY(wechatFileHelper, "https://filehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsg?lang=zh_CN&pass_ticket="+url.QueryEscape(wechatFileHelper.PassTicket), string(marshal))), &messageSentRetBody)
		if err != nil {
			panic(err)
		}
		if messageSentRetBody.BaseResponse.Ret != 0 {
			println("发送消息失败！")
		}
	}

}

// SendImgMessage 文件传输发动图片信息
func (wechatFileHelper *WechatFileHelper) SendImgMessage() {
	//略微麻烦，暂时不做了。
	//先上传
	//https://file.wx.qq.com/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json&random=Xxdi
	//form data:
	//name: 微信图片_20220517152257.jpg
	//lastModifiedDate: Tue May 17 2022 15:22:56 GMT+0800 (中国标准时间)
	//size: 19838
	//type: image/jpeg
	//mediatype: pic
	//uploadmediarequest: {"UploadType":2,"BaseRequest":{"Uin":844282903,"Sid":"SNa6sRlEHdaQDq5W","Skey":"@crypt_e7a64693_e205fe286c2655ed172a912100843024","DeviceID":"382490164959149"},"ClientMediaId":"16615960516710466","TotalLen":19838,"StartPos":0,"DataLen":19838,"MediaType":4,"FromUserName":"@fd26181d7fe3c5ef32a2c20324ee131c378cdb65bf4d321da074dea52526c7dc","ToUserName":"filehelper","FileMd5":"c3055e653564d4c25762fd5cc41d5eba"}
	//webwx_data_ticket: gScW0JaoQNDdT9DvkJXnZDuC
	//pass_ticket: 1OiOiNu8V6E%2BB2rxI4Nms5lpc2Uf7xZz5wQQN7ACSgvOYncm5K%2FpjOCKWqAvX4z%2B
	//filename: （二进制）

	//Respond:
	//{
	//"BaseResponse": {
	//"Ret": 0,
	//"ErrMsg": ""
	//}
	//,
	//"MediaId": "@crypt_e3561f0e_3c8145f5a1e7341794a65092e74a1cd9a8bc786fb2075cbb564ab1337cd0233d7ded239b1583ae99496ed4944f19fc60fdc4380110e9da76a3a08b26b162142fb76d6159437207385cbe68a62d1ad11026f2fe4416f742ace32a14fc0b03c626ab7c9a4b528337ca72fa16ea329fa6d4af69ea789a6037d752ef7b82a98a651b482ce0080fe7e37d40ca2b979109bedd5cc5c2c0df385ba23f5052987e9f332b9531a8023bdc3342085a5bbbaca272202b006b10356a6c9a68b6eebe1086216d610a7464e3479846870d248d22cf6f4df51623f4246e7abeb5ae60e13a045c1eee4f88217a1c6debb73ec6e093688ec0e2fe69b1c4d57c47f79f5b5de1bb6fa784d77ea39d136b786081c5c259b9008cf81a48d494a5a49f292ba0944069b484",
	//"StartPos": 19838,
	//"CDNThumbImgHeight": 100,
	//"CDNThumbImgWidth": 100,
	//"EncryFileName": "%E5%BE%AE%E4%BF%A1%E5%9B%BE%E7%89%87%5F20220517152257%2Ejpg"
	//}

	//发送
	//https://filehelper.weixin.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsgimg?fun=async&f=json&pass_ticket=1OiOiNu8V6E%2BB2rxI4Nms5lpc2Uf7xZz5wQQN7ACSgvOYncm5K%2FpjOCKWqAvX4z%2B
	//POST
	//{"BaseRequest":{"Uin":844282903,"Sid":"SNa6sRlEHdaQDq5W","Skey":"@crypt_e7a64693_e205fe286c2655ed172a912100843024","DeviceID":"810830073576311"},"Msg":{"ClientMsgId":"16615960579560994","FromUserName":"@fd26181d7fe3c5ef32a2c20324ee131c378cdb65bf4d321da074dea52526c7dc","LocalID":"16615960579560994","ToUserName":"filehelper","MediaId":"@crypt_e3561f0e_3c8145f5a1e7341794a65092e74a1cd9a8bc786fb2075cbb564ab1337cd0233d7ded239b1583ae99496ed4944f19fc60fdc4380110e9da76a3a08b26b162142fb76d6159437207385cbe68a62d1ad11026f2fe4416f742ace32a14fc0b03c626ab7c9a4b528337ca72fa16ea329fa6d4af69ea789a6037d752ef7b82a98a651b482ce0080fe7e37d40ca2b979109bedd5cc5c2c0df385ba23f5052987e9f332b9531a8023bdc3342085a5bbbaca272202b006b10356a6c9a68b6eebe1086216d610a7464e3479846870d248d22cf6f4df51623f4246e7abeb5ae60e13a045c1eee4f88217a1c6debb73ec6e093688ec0e2fe69b1c4d57c47f79f5b5de1bb6fa784d77ea39d136b786081c5c259b9008cf81a48d494a5a49f292ba0944069b484","Type":3,"Content":""},"Scene":0}
}

// GetSZFileBaseRequest SzFile版本的GetBaseRequest
func (wechatFileHelper *WechatFileHelper) GetSZFileBaseRequest() SzFileBaseRequestBody {
	return SzFileBaseRequestBody{SzBaseRequest: struct {
		Uin      string `json:"Uin"`
		Sid      string `json:"Sid"`
		Skey     string `json:"Skey"`
		DeviceId string `json:"DeviceId"`
	}(struct {
		Uin      string
		Sid      string
		Skey     string
		DeviceId string
	}{DeviceId: GetRandomString(15),
		Sid:  wechatFileHelper.WxSid,
		Skey: wechatFileHelper.SKey,
		Uin:  strconv.Itoa(wechatFileHelper.WxUin)})}

}

// GetBaseRequest 老版的微信文件助手的BaseRequest请求。
func (wechatFileHelper *WechatFileHelper) GetBaseRequest() BaseRequestBody {
	return BaseRequestBody{BaseRequest: struct {
		Uin      int    `json:"Uin"`
		Sid      string `json:"Sid"`
		Skey     string `json:"Skey"`
		DeviceId string `json:"DeviceId"`
	}(struct {
		Uin      int
		Sid      string
		Skey     string
		DeviceId string
	}{DeviceId: GetRandomString(15),
		Sid:  wechatFileHelper.WxSid,
		Skey: wechatFileHelper.SKey,
		Uin:  wechatFileHelper.WxUin})}

}
