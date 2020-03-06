package bilibili_tools_go

import (
	"net/http"
)

// Bilibili is a struct for easy net.client access
type Bilibili struct {
	client *http.Client
	info   *loginInfo
}

type loginInfo struct {
	UserName     string `json:"user_name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Cookies      string `json:"cookies"`
}

type userInfo struct {
	Code   int  `json:"code"`
	Status bool `json:"status"`
	Data   struct {
		LevelInfo struct {
			CurrentLevel int `json:"current_level"`
			CurrentMin   int `json:"current_min"`
			CurrentExp   int `json:"current_exp"`
			NextExp      int `json:"next_exp"`
		} `json:"level_info"`
		BCoins           int     `json:"bCoins"`
		Coins            float64 `json:"coins"`
		Face             string  `json:"face"`
		NameplateCurrent string  `json:"nameplate_current"`
		UName            string  `json:"uname"`
		UserStatus       string  `json:"userStatus"`
		VipType          int     `json:"vipType"`
		VipStatus        int     `json:"vipStatus"`
		OfficialVerify   int     `json:"official_verify"`
	} `json:"data"`
}

type rsaLogin struct {
	Code int `json:"code"`
	Data struct {
		Hash string `json:"hash"`
		Key  string `json:"key"`
	} `json:"data"`
}

type userAccess struct {
	Code int `json:"code"`
	Data struct {
		Status    int `json:"status"`
		TokenInfo struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"token_info"`
		CookieInfo struct {
			Cookies []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"cookies"`
		} `json:"cookie_info"`
	} `json:"data"`
	Message string `json:"message"`
}

type liveReceivedGiftList struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Data    struct {
		List []struct {
			UID      int    `json:"uid"`
			UName    string `json:"uname"`
			Time     string `json:"time"`
			GiftID   int    `json:"gift_id"`
			GiftName string `json:"gift_name"`
			GiftNum  int    `json:"gift_num"`
			Hamster  int    `json:"hamster"`
		} `json:"list"`
		HasMore    int    `json:"has_more"`
		NextOffset string `json:"next_offset"`
	} `json:"data"`
}
