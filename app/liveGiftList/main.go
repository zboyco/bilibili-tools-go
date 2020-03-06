package main

import (
	"fmt"
	"github.com/zboyco/bilibili-tools-go"
)

func main() {
	defer func() {
		var exit string
		fmt.Println("\n按 回车 键退出...")
		fmt.Scanln(&exit)
	}()

	bili, err := bilibili_tools_go.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	var date string
	fmt.Print("请输入需要导出的日期(如:20200305): ")
	fmt.Scanln(&date)
	bili.DownloadReceivedGiftList(date)
}
