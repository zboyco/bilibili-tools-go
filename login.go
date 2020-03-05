package bilibili_tools_go

import (
	"errors"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// NewFromLogin 账号密码登录
func NewFromLogin(username, password string) (*Bilibili, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	ret := &Bilibili{Client: &http.Client{Jar: jar}}
	if err = ret.login(username, password); err != nil {
		return nil, err
	}
	return ret, nil
}

// NewFromCookie Cookie登录
func NewFromCookie(cookie string) (*Bilibili, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	procCookie := stringToCookie(cookie)
	for i := range procCookie {
		procCookie[i].Domain = ".bilibili.com"
		procCookie[i].Path = "/"
	}
	jar.SetCookies(BiliLoginURL, procCookie)
	ret := &Bilibili{Client: &http.Client{Jar: jar}}
	return ret, nil
}

func (bili *Bilibili) login(username, password string) error {
	// 请求首页和登录页
	if _, err := http.Get(MainHost); err != nil {
		return err
	}
	if _, err := http.Get(LoginHost); err != nil {
		return err
	}

	// 获取加密密码串
	pass, err := rsaEncryptPwd(password)
	if err != nil {
		return err
	}

	// 构建登录请求
	data := url.Values{}
	data.Add("appkey", AppKey)
	data.Add("username", username)
	data.Add("password", pass)
	encode := data.Encode()
	payload := fmt.Sprintf("%s&sign=%s", encode, calcSign(encode))

	req, err := network(LoginUrl, "POST", payload)
	if err != nil {
		return err
	}
	resp, err := bili.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	user := &userAccess{}
	if err = jsonProc(resp, user); err != nil {
		return err
	}

	if user.Code == -449 {
		return errors.New("retry later")
	}
	if user.Code == -105 {
		// 验证码
	}
	if user.Code == 0 && user.Data.Status == 0 {
		// 正常
		cookies := make([]*http.Cookie, len(user.Data.CookieInfo.Cookies))
		for i, v := range user.Data.CookieInfo.Cookies {
			cookies[i] = &http.Cookie{
				Name:   v.Name,
				Value:  v.Value,
				Domain: ".bilibili.com",
				Path:   "/",
			}
		}
		bili.Client.Jar.SetCookies(BiliLoginURL, cookies)
		return nil
	}
	return fmt.Errorf("Unknown error with code: %d", user.Code)
}

// IsLoggedIn 判断是否登录成功
func (bili *Bilibili) IsLoggedIn() (bool, error) {
	req, err := network("https://account.bilibili.com/home/userInfo", "GET", "")
	if err != nil {
		return false, err
	}
	resp, err := bili.Client.Do(req)
	userInfo := new(userInfo)
	if err = jsonProc(resp, userInfo); err != nil {
		return false, err
	}
	return userInfo.Code == 0, nil
}

// 获取验证码
func getCaptcha(client *http.Client) (string, error) {
	var ret string
	req, err := network("https://passport.bilibili.com/captcha", "GET", "")
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", "https://passport.bilibili.com/ajax/miniLogin/minilogin")
	req.Header.Set("Accept", "image/jpeg")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tmpJPG := filepath.Join(os.TempDir(), "vdcode.jpg")
	tmpFile, err := os.Create(tmpJPG)
	if err != nil {
		return "", err
	}
	defer syscall.Unlink(tmpJPG)

	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}
	tmpFile.Close()

	err = open.Start(tmpJPG)
	if err == nil {
		fmt.Print("请输入你看到的验证码并回车：")
	} else {
		fmt.Printf("打开图片失败，请自行打开%s，输入验证码并回车：", tmpJPG)
	}
	if _, err = fmt.Scanf("%s", &ret); err != nil {
		return "", err
	}
	return strings.ToLower(ret), nil
}


