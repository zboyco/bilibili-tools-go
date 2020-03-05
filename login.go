package bilibili_tools_go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// NewFromLogin 账号密码登录
func NewFromLogin(userName, userPwd string) (*Bilibili, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	bili := &Bilibili{Client: &http.Client{Jar: jar}}
	if err = bili.login(userName, userPwd, ""); err != nil {
		return nil, err
	}
	return bili, nil
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

func (bili *Bilibili) login(userName, userPwd, captcha string) error {
	// 请求首页和登录页
	if _, err := http.Get(MainHost); err != nil {
		return err
	}
	if _, err := http.Get(LoginHost); err != nil {
		return err
	}

	// 获取加密密码串
	pass, err := rsaEncryptPwd(userPwd)
	if err != nil {
		return err
	}

	// 构建登录请求
	data := url.Values{}
	data.Add("appkey", AppKey)
	data.Add("username", userName)
	data.Add("password", pass)
	if captcha != "" {
		data.Add("captcha", captcha)
	}
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
		fmt.Println("登录失败，需要验证码，正在获取验证码...")
		code, err := bili.getCaptcha()
		if err != nil {
			return err
		}
		return bili.login(userName, userPwd, code)
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
	return fmt.Errorf("error with code: %d, %s", user.Code, user.Message)
}

// IsLoggedIn 判断是否登录成功
func (bili *Bilibili) IsLoggedIn() (bool, error) {
	req, err := network("https://account.bilibili.com/home/userInfo", "GET", "")
	if err != nil {
		return false, err
	}
	resp, err := bili.Client.Do(req)
	defer resp.Body.Close()
	userInfo := new(userInfo)
	if err = jsonProc(resp, userInfo); err != nil {
		return false, err
	}
	return userInfo.Code == 0, nil
}

// 获取登录验证码
func (bili *Bilibili) getCaptcha() (string, error) {
	var ret string
	req, err := network("https://passport.bilibili.com/captcha", "GET", "")
	if err != nil {
		return "", err
	}
	req.Header.Set("Host", "passport.bilibili.com")
	req.Header.Set("Referer", "https://passport.bilibili.com/login")
	req.Header.Set("Accept", "image/jpeg")
	resp, err := bili.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imagesData, _ := ioutil.ReadAll(resp.Body)
	ret, err = identifyCaptcha(imagesData)
	if err == nil {
		fmt.Println("自动识别验证码成功...")
		return strings.ToUpper(ret), nil
	}
	fmt.Println("自动识别验证码失败，请手动填写...")

	tmpFilePath := filepath.Join(os.TempDir(), "bilibili-captcha.jpg")
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return "", err
	}
	// 删除临时图片
	defer syscall.Unlink(tmpFilePath)

	if _, err = tmpFile.Write(imagesData); err != nil {
		return "", err
	}
	tmpFile.Close()

	fmt.Println("正在打开验证码图片...")
	err = open.Start(tmpFilePath)
	if err != nil {
		fmt.Printf("打开图片失败，请自行打开【%s】\n", tmpFilePath)
	}
	fmt.Print("请输入验证码并回车：")
	if _, err = fmt.Scanf("%s", &ret); err != nil {
		return "", err
	}
	return strings.ToUpper(ret), nil
}

// 识别验证码
func identifyCaptcha(src []byte) (string, error) {
	body := make(map[string]string)
	body["image"] = base64.StdEncoding.EncodeToString(src)
	bytesData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("https://bili.dev:2233/captcha", "application/json", bytes.NewBuffer(bytesData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result := &struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Success bool   `json:"success"`
	}{}
	if err = jsonProc(resp, result); err != nil {
		return "", err
	}
	if result.Code != 0 {
		return "", errors.New(result.Message)
	}
	return result.Message, nil
}
