package bilibili_tools_go

import (
	"net/url"
)

const (
	UserAgent = "Mozilla/5.0 BiliDroid/5.51.1 (xxxxxx@gmail.com)"
	MainHost  = "https://www.bilibili.com"
	LoginHost = "https://passport.bilibili.com/login"
	LoginUrl  = "https://passport.bilibili.com/api/v2/oauth2/login"

	AppKey    = "1d8b6e7d45233436"
	SecretKey = "560c52ccd288fed045859ed18bffd973"

	LiveReceivedGiftList = "https://api.live.bilibili.com/gift/v1/master/getReceivedGiftList"
)

var BiliLoginURL *url.URL

func init() {
	BiliLoginURL, _ = url.Parse(LoginUrl)
}
