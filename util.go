package bilibili_tools_go

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func network(url, method, query string) (req *http.Request, err error) {
	switch method {
	case "GET":
		req, err = http.NewRequest("GET", url, nil)
		req.URL.RawQuery = query
	case "POST":
		req, err = http.NewRequest("POST", url, strings.NewReader(query))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8;")
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	return
}

func jsonProc(body *http.Response, container interface{}) error {
	defer body.Body.Close()
	if err := json.NewDecoder(body.Body).Decode(container); err != nil {
		return err
	}
	return nil
}

func calcSign(param string) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s%s", param, SecretKey)))
	return hex.EncodeToString(h.Sum(nil))
}

func rsaEncryptPwd(password string) (string, error) {
	ret := &rsaLogin{}
	payload := fmt.Sprintf("appkey=%s&sign=%s", AppKey, calcSign(fmt.Sprintf("appkey=%s", AppKey)))
	resp, err := http.Post(OAuth2GetKeyUrl, "application/x-www-form-urlencoded; charset=utf-8;", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	if err = jsonProc(resp, ret); err != nil {
		return "", err
	}
	crypt, err := rsaEncrypt([]byte(ret.Data.Key), []byte(ret.Data.Hash+password))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(crypt), nil
}

func rsaEncrypt(publicKey []byte, origData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
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
