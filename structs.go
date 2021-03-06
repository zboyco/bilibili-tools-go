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
	UserName     string            `json:"user_name"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	Cookies      map[string]string `json:"cookies"`
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

type oAuth2Login struct {
	Code int `json:"code"`
	TS   int `json:"ts"`
	Data struct {
		Status    int `json:"status"`
		TokenInfo struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		} `json:"token_info"`
		CookieInfo struct {
			Cookies []struct {
				Name    string `json:"name"`
				Value   string `json:"value"`
				Expires int    `json:"expires"`
			} `json:"cookies"`
		} `json:"cookie_info"`
	} `json:"data"`
	Message string `json:"message"`
}

type oAuth2Refresh struct {
	Code int `json:"code"`
	TS   int `json:"ts"`
	Data struct {
		TokenInfo struct {
			MID          int    `json:"mid"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		} `json:"token_info"`
		CookieInfo struct {
			Cookies []struct {
				Name    string `json:"name"`
				Value   string `json:"value"`
				Expires int    `json:"expires"`
			} `json:"cookies"`
		} `json:"cookie_info"`
	} `json:"data"`
	Message string `json:"message"`
}

// 直播心跳
type liveHeartBeat struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Data    struct {
		Open   int `json:"open"`
		HasNew int `json:"has_new"`
		Count  int `json:"count"`
	} `json:"data"`
}

// 直播礼物流水
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

// 直播签到
type liveDoSign struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Text        string `json:"text"`
		SpecialText string `json:"specialText"`
		AllDays     int    `json:"allDays"`
		HadSignDays int    `json:"hadSignDays"`
		IsBonusDay  int    `json:"isBonusDay"`
	} `json:"data"`
}

// 直播银瓜子换硬币
type silver2Coin struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Data    struct {
		Gold   string `json:"gold"`
		Silver string `json:"silver"`
		TID    string `json:"tid"`
		Coin   int    `json:"coin"`
	} `json:"data"`
}
