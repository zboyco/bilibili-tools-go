package bilibili_tools_go

import (
	"bufio"
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
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Login() (*Bilibili, error) {
	var (
		userName, userPwd string
		err               error
	)

	// 读取用户账号密码
	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("请输入账号: ")
	if reader.Scan() {
		userName = reader.Text()
	}

	// 新建对象
	jar, _ := cookiejar.New(nil)
	bili := &Bilibili{client: &http.Client{Jar: jar}, info: &loginInfo{UserName: userName}}

	// 读取本地登录信息
	info := readLoginInfo(userName)
	if info != nil {
		fmt.Println("尝试缓存登录...")
		bili.info = info
		// cookie
		err = bili.byCookie()
		if err == nil {
			fmt.Println("登录成功...")
			return bili, nil
		}
		// token
		err = bili.byToken()
		if err == nil {
			bili.saveLoginInfo()
			fmt.Println("登录成功...")
			return bili, nil
		}
		fmt.Println("缓存登录失败...")
	}

	fmt.Print("请输入密码: ")
	if reader.Scan() {
		userPwd = reader.Text()
	}
	err = bili.byPassword(userName, userPwd, "")
	if err != nil {
		return nil, err
	}
	bili.saveLoginInfo()
	fmt.Println("登录成功...")
	return bili, nil
}

// 读取配置
func readLoginInfo(userName string) *loginInfo {
	exist, _ := pathExists(ConfigFileName)
	if !exist {
		return nil
	}
	jsonByte, err := ioutil.ReadFile(ConfigFileName)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	users := map[string]*loginInfo{}
	err = json.Unmarshal(jsonByte, &users)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if _, exist := users[userName]; !exist {
		return nil
	}
	return users[userName]
}

// Cookie登录
func (bili *Bilibili) byCookie() error {
	cookies := make([]*http.Cookie, 0)
	for key, v := range bili.info.Cookies {
		cookies = append(cookies, &http.Cookie{
			Name:   key,
			Value:  v,
			Domain: ".bilibili.com",
			Path:   "/",
		})
	}
	bili.client.Jar.SetCookies(BiliLoginURL, cookies)
	logged, err := bili.IsLoggedIn()
	if err != nil {
		return err
	}
	if !logged {
		return errors.New("by cookie fail")
	}
	return nil
}

// token登录
func (bili *Bilibili) byToken() error {
	// 获取cookie
	data := url.Values{}
	data.Add("access_key", bili.info.AccessToken)
	data.Add("appkey", AppKey)
	data.Add("refresh_token", bili.info.RefreshToken)
	data.Add("ts", strconv.FormatInt(time.Now().Unix(), 10))
	encode := data.Encode()
	payload := fmt.Sprintf("%s&sign=%s", encode, calcSign(encode))
	req, err := network(OAuth2RefreshUrl, "POST", payload)
	if err != nil {
		return err
	}
	resp, err := bili.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	result := &oAuth2Refresh{}
	if err = jsonProc(resp, result); err != nil {
		return err
	}
	if result.Code != 0 {
		return errors.New(result.Message)
	}
	// 正常
	bili.info.Cookies = make(map[string]string)
	cookies := make([]*http.Cookie, len(result.Data.CookieInfo.Cookies))
	for i, v := range result.Data.CookieInfo.Cookies {
		bili.info.Cookies[v.Name] = v.Value
		cookies[i] = &http.Cookie{
			Name:   v.Name,
			Value:  v.Value,
			Domain: ".bilibili.com",
			Path:   "/",
		}
	}
	bili.info.AccessToken = result.Data.TokenInfo.AccessToken
	bili.info.RefreshToken = result.Data.TokenInfo.RefreshToken
	bili.client.Jar.SetCookies(BiliLoginURL, cookies)
	return nil
}

// 账号密码登录
func (bili *Bilibili) byPassword(userName, userPwd, captcha string) error {
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

	req, err := network(OAuth2LoginUrl, "POST", payload)
	if err != nil {
		return err
	}
	resp, err := bili.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	user := &oAuth2Login{}
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
		return bili.byPassword(userName, userPwd, code)
	}
	if user.Code == 0 && user.Data.Status == 0 {
		// 正常
		bili.info.Cookies = make(map[string]string)
		cookies := make([]*http.Cookie, len(user.Data.CookieInfo.Cookies))
		for i, v := range user.Data.CookieInfo.Cookies {
			bili.info.Cookies[v.Name] = v.Value
			cookies[i] = &http.Cookie{
				Name:   v.Name,
				Value:  v.Value,
				Domain: ".bilibili.com",
				Path:   "/",
			}
		}
		bili.info.AccessToken = user.Data.TokenInfo.AccessToken
		bili.info.RefreshToken = user.Data.TokenInfo.RefreshToken
		bili.client.Jar.SetCookies(BiliLoginURL, cookies)
		return nil
	}
	return fmt.Errorf("error with code: %d, %s", user.Code, user.Message)
}

// IsLoggedIn 判断是否登录成功
func (bili *Bilibili) IsLoggedIn() (bool, error) {
	req, err := network(UserInfoUrl, "GET", "")
	if err != nil {
		return false, err
	}
	resp, err := bili.client.Do(req)
	defer resp.Body.Close()
	userInfo := &userInfo{}
	if err = jsonProc(resp, userInfo); err != nil {
		return false, err
	}
	return userInfo.Code == 0, nil
}

// 获取登录验证码
func (bili *Bilibili) getCaptcha() (string, error) {
	var ret string
	req, err := network(CaptchaUrl, "GET", "")
	if err != nil {
		return "", err
	}
	req.Header.Set("Host", "passport.bilibili.com")
	req.Header.Set("Referer", LoginHost)
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

// 保存登录信息
func (bili *Bilibili) saveLoginInfo() {
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
		_, exist := users[bili.info.UserName]
		if exist {
			delete(users, bili.info.UserName)
		}
	}
	users[bili.info.UserName] = bili.info

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
