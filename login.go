package bilibili_tools_go

import (
	"bufio"
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

func Login() (*Bilibili, error) {
	var userName, userPwd string
	// 读取用户账号密码
	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("请输入账号: ")
	if reader.Scan() {
		userName = reader.Text()
	}

	fmt.Println("尝试缓存登录...")
	bili, err := byCookie(userName)
	if err == nil {
		fmt.Println("登录成功...")
		return bili, nil
	}

	fmt.Print("缓存登录失败，请输入密码: ")
	if reader.Scan() {
		userPwd = reader.Text()
	}
	bili, err = byPassword(userName, userPwd)
	if err != nil {
		return nil, err
	}

	fmt.Println("登录成功...")
	return bili, nil
}

// 账号密码登录
func byPassword(userName, userPwd string) (*Bilibili, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	bili := &Bilibili{client: &http.Client{Jar: jar}, info: &loginInfo{UserName: userName}}
	if err = bili.login(userName, userPwd, ""); err != nil {
		return nil, err
	}
	saveLoginInfo(bili.info)
	return bili, nil
}

// Cookie登录
func byCookie(userName string) (*Bilibili, error) {
	exist, _ := pathExists(ConfigFileName)
	if !exist {
		return nil, errors.New("no config")
	}
	jsonByte, err := ioutil.ReadFile(ConfigFileName)
	if err != nil {
		return nil, err
	}

	users := map[string]*loginInfo{}
	err = json.Unmarshal(jsonByte, &users)
	if err != nil {
		return nil, err
	}
	if _, exist := users[userName]; !exist {
		return nil, errors.New("no config")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	procCookie := stringToCookie(users[userName].Cookies)
	for i := range procCookie {
		procCookie[i].Domain = ".bilibili.com"
		procCookie[i].Path = "/"
	}
	jar.SetCookies(BiliLoginURL, procCookie)
	ret := &Bilibili{client: &http.Client{Jar: jar}, info: users[userName]}
	logged, err := ret.IsLoggedIn()
	if err != nil {
		return nil, err
	}
	if !logged {
		return nil, errors.New("by cookie fail")
	}
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
	resp, err := bili.client.Do(req)
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
		var cookiesStr string
		cookies := make([]*http.Cookie, len(user.Data.CookieInfo.Cookies))
		for i, v := range user.Data.CookieInfo.Cookies {
			cookiesStr += fmt.Sprintf("%s=%s; ", v.Name, v.Value)
			cookies[i] = &http.Cookie{
				Name:   v.Name,
				Value:  v.Value,
				Domain: ".bilibili.com",
				Path:   "/",
			}
		}
		bili.info.Cookies = cookiesStr
		bili.info.AccessToken = user.Data.TokenInfo.AccessToken
		bili.info.RefreshToken = user.Data.TokenInfo.RefreshToken
		bili.client.Jar.SetCookies(BiliLoginURL, cookies)
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
	resp, err := bili.client.Do(req)
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
	resp, err := bili.client.Do(req)
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

// 保存登录信息
func saveLoginInfo(info *loginInfo) {
	f, err := os.OpenFile(ConfigFileName, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonByte, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	users := map[string]*loginInfo{}

	if len(jsonByte) > 0 {
		if err := json.Unmarshal(jsonByte, &users); err != nil {
			fmt.Println(err)
			return
		}
		_, exist := users[info.UserName]
		if exist {
			delete(users, info.UserName)
		}
	}
	users[info.UserName] = info

	jsonByte, err = json.MarshalIndent(users, "", "	")
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Truncate(0)
	f.Seek(0, 0)
	_, err = f.Write(jsonByte)
	if err != nil {
		fmt.Println(err)
	}
}
