package _interface

import "encoding/xml"

//微信登录相关结构体
type Webwxnewloginpage struct {
	XMLName     xml.Name `xml:"error"`
	Ret         int      `xml:"ret"`
	Message     string   `xml:"message"`
	Skey        string   `xml:"skey"`
	WxSid       string   `xml:"wxsid"`
	WxUin       int      `xml:"wxuin"`
	PassTicket  string   `xml:"pass_ticket"`
	IsGrayScale int      `xml:"isgrayscale"`
}

type BaseRequestBody struct {
	BaseRequest struct {
		Uin      int    `json:"Uin"`
		Sid      string `json:"Sid"`
		Skey     string `json:"Skey"`
		DeviceId string `json:"DeviceId"`
	} `json:"BaseRequest"`
}

type WebwxInitRetBody struct {
	SyncKey struct {
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
}

type SyncCheckBody struct {
	Retcode  string `json:"retcode"`
	Selector string `json:"selector"`
}

//MsgType
//1 = text
//3 = img
//34 = voice
//43 = video
type SyncBody struct {
	BaseResponse struct {
		Ret    int    `json:"Ret"`
		ErrMsg string `json:"ErrMsg"`
	} `json:"BaseResponse"`
	AddMsgCount int `json:"AddMsgCount"`
	AddMsgList  []struct {
		MsgId                string `json:"MsgId"`
		FromUserName         string `json:"FromUserName"`
		ToUserName           string `json:"ToUserName"`
		MsgType              int    `json:"MsgType"`
		Content              string `json:"Content"`
		Status               int    `json:"Status"`
		ImgStatus            int    `json:"ImgStatus"`
		CreateTime           int    `json:"CreateTime"`
		VoiceLength          int    `json:"VoiceLength"`
		PlayLength           int    `json:"PlayLength"`
		FileName             string `json:"FileName"`
		FileSize             string `json:"FileSize"`
		MediaId              string `json:"MediaId"`
		Url                  string `json:"Url"`
		AppMsgType           int    `json:"AppMsgType"`
		StatusNotifyCode     int    `json:"StatusNotifyCode"`
		StatusNotifyUserName string `json:"StatusNotifyUserName"`
		ForwardFlag          int    `json:"ForwardFlag"`
		AppInfo              struct {
			AppID string `json:"AppID"`
			Type  int    `json:"Type"`
		} `json:"AppInfo"`
		HasProductId  int    `json:"HasProductId"`
		Ticket        string `json:"Ticket"`
		ImgHeight     int    `json:"ImgHeight"`
		ImgWidth      int    `json:"ImgWidth"`
		SubMsgType    int    `json:"SubMsgType"`
		NewMsgId      int64  `json:"NewMsgId"`
		OriContent    string `json:"OriContent"`
		EncryFileName string `json:"EncryFileName"`
	} `json:"AddMsgList"`
	SyncKey struct {
		Count int `json:"Count"`
		List  []struct {
			Key int `json:"Key"`
			Val int `json:"Val"`
		} `json:"List"`
	} `json:"SyncKey"`
}

type XsyncBody struct {
	BaseRequest struct {
		Uin      int    `json:"Uin"`
		Sid      string `json:"Sid"`
		Skey     string `json:"Skey"`
		DeviceId string `json:"DeviceId"`
	} `json:"BaseRequest"`
	SyncKey struct {
		Count int `json:"Count"`
		List  []struct {
			Key int `json:"Key"`
			Val int `json:"Val"`
		} `json:"List"`
	} `json:"SyncKey"`
	Rr int `json:"rr"`
}

type SendTextBody struct {
	BaseRequest struct {
		Uin      int    `json:"Uin"`
		Sid      string `json:"Sid"`
		Skey     string `json:"Skey"`
		DeviceID string `json:"DeviceID"`
	} `json:"BaseRequest"`
	Msg struct {
		ClientMsgId  string `json:"ClientMsgId"`
		FromUserName string `json:"FromUserName"`
		LocalID      string `json:"LocalID"`
		ToUserName   string `json:"ToUserName"`
		Content      string `json:"Content"`
		Type         int    `json:"Type"`
	} `json:"Msg"`
	Scene int `json:"Scene"`
}

type MessageSentRetBody struct {
	BaseResponse struct {
		Ret    int    `json:"Ret"`
		ErrMsg string `json:"ErrMsg"`
	} `json:"BaseResponse"`
	MsgID   string `json:"MsgID"`
	LocalID string `json:"LocalID"`
}
