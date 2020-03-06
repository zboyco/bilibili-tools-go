package main

import (
	"fmt"
	"github.com/zboyco/bilibili-tools-go"
)

func main() {
	defer func() {
		var in string
		fmt.Println("按 回车 键退出...")
		fmt.Scanln(&in)
	}()

	bili, err := bilibili_tools_go.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	var date string
	fmt.Print("请输入需要导出的日期(如:20200305): ")
	fmt.Scanf("%s", &date)
	bili.DownloadReceivedGiftList(date)
}
