package main

import (
	"bufio"
	"fmt"
	"github.com/zboyco/bilibili-tools-go"
	"os"
)

func main() {
	defer func() {
		var in string
		fmt.Println("按 回车 键退出...")
		fmt.Scanln(&in)
	}()

	var userName, userPwd string
	// 读取用户账号密码
	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("请输入账号: ")
	if reader.Scan() {
		userName = reader.Text()
	}
	fmt.Print("请输入密码: ")
	if reader.Scan() {
		userPwd = reader.Text()
	}
	fmt.Println("尝试登录...")
	b, err := bilibili_tools_go.NewFromLogin(userName, userPwd)
	if err != nil {
		fmt.Println(err)
		return
	}
	logged, err := b.IsLoggedIn()
	if err != nil {
		fmt.Println("登录失败：", err)
		return
	}
	if !logged {
		fmt.Println("登录失败：稍后重试！")
		return
	}

	fmt.Println("登录成功...", )
	var date string
	fmt.Print("请输入需要导出的日期(如:20200305): ")
	if reader.Scan() {
		date = reader.Text()
	}
	b.DownloadReceivedGiftList(date)
}
