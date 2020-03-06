package bilibili_tools_go

import (
	"net/url"
)

const (
	ConfigFileName = "config.json"

	UserAgent        = "Mozilla/5.0 BiliDroid/5.51.1 (xxxxxx@gmail.com)"
	MainHost         = "https://www.bilibili.com"
	LoginHost        = "https://passport.bilibili.com/login"
	CaptchaUrl       = "https://passport.bilibili.com/captcha"
	OAuth2LoginUrl   = "https://passport.bilibili.com/api/v2/oauth2/login"
	OAuth2GetKeyUrl  = "https://passport.bilibili.com/api/oauth2/getKey"
	OAuth2RefreshUrl = "https://passport.bilibili.com/api/v2/oauth2/refresh_token"

	AppKey    = "1d8b6e7d45233436"
	SecretKey = "560c52ccd288fed045859ed18bffd973"

	UserInfoUrl             = "https://account.bilibili.com/home/userInfo"
	LiveReceivedGiftListUrl = "https://api.live.bilibili.com/gift/v1/master/getReceivedGiftList"
)

var BiliLoginURL *url.URL

func init() {
	BiliLoginURL, _ = url.Parse(OAuth2LoginUrl)
}
