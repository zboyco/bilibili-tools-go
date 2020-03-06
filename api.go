package bilibili_tools_go

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// DownloadReceivedGiftList 下载金瓜子礼物流水
func (bili *Bilibili) DownloadReceivedGiftList(date string) {
	// 访问直播中心礼物流水页面
	_, err := bili.Client.Get("https://link.bilibili.com/p/center/index#/my-room/gift-list")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("开始导出...")

	// 创建文件
	f := excelize.NewFile()
	_ = f.SetCellValue("Sheet1", "A1", "用户昵称")
	_ = f.SetCellValue("Sheet1", "B1", "用户ID")
	_ = f.SetCellValue("Sheet1", "C1", "礼物名称")
	_ = f.SetCellValue("Sheet1", "D1", "数量")
	_ = f.SetCellValue("Sheet1", "E1", "金仓鼠")
	_ = f.SetCellValue("Sheet1", "F1", "时间")

	var wg sync.WaitGroup
	wg.Add(2)
	writeQueue := make(chan *liveReceivedGiftList, 10)

	go func(q <-chan *liveReceivedGiftList) {
		defer wg.Done()
		// 写入excel
		line := 1
		for {
			m, ok := <-q
			if !ok {
				return
			}
			for _, item := range m.Data.List {
				line++
				_ = f.SetCellValue("Sheet1", fmt.Sprintf("A%v", line), item.UName)
				_ = f.SetCellValue("Sheet1", fmt.Sprintf("B%v", line), strconv.Itoa(item.UID))
				_ = f.SetCellValue("Sheet1", fmt.Sprintf("C%v", line), item.GiftName)
				_ = f.SetCellValue("Sheet1", fmt.Sprintf("D%v", line), item.GiftNum)
				_ = f.SetCellValue("Sheet1", fmt.Sprintf("E%v", line), item.Hamster)
				_ = f.SetCellValue("Sheet1", fmt.Sprintf("F%v", line), item.Time)
				fmt.Println(item.UName, ":", item.GiftName, "*", item.GiftNum)
			}
		}
	}(writeQueue)

	go func(q chan<- *liveReceivedGiftList) {
		defer wg.Done()
		defer close(q)
		// 从接口获取数据
		offset := ""
		hasMore := 1
		for hasMore == 1 {
			time.Sleep(time.Duration(rand.Intn(600)+400) * time.Millisecond)

			list, err := bili.getReceivedGiftList(date, offset)
			if err != nil {
				fmt.Println(err)
				return
			}

			if list.Code != 0 {
				fmt.Println(list.Message)
				return
			}

			hasMore = list.Data.HasMore
			offset = list.Data.NextOffset

			q <- list
		}
	}(writeQueue)

	wg.Wait()
	// 保存文件
	fileName := fmt.Sprintf("%v礼物流水-%v.xlsx", date, time.Now().Format("20060102150405"))
	fmt.Println("")
	if err := f.SaveAs(fileName); err != nil {
		fmt.Println("保存失败，", err.Error())
	} else {
		fmt.Println("读取完成，数据已保存到", fileName)
	}
}

func (bili *Bilibili) getReceivedGiftList(date, offset string) (*liveReceivedGiftList, error) {
	// 构建登录请求
	data := url.Values{}
	data.Add("coin_type", "gold")
	data.Add("gift_id", "")
	data.Add("begin_time", date)
	data.Add("uname", "")
	data.Add("next_offset", offset)
	data.Add("page_size", "20")

	req, err := network(LiveReceivedGiftList, "GET", data.Encode())
	req.Header.Set("Host", "api.live.bilibili.com")
	req.Header.Set("Origin", "https://link.bilibili.com")
	req.Header.Set("Referer", "https://link.bilibili.com/p/center/index")
	if err != nil {
		return nil, err
	}
	resp, err := bili.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	user := &liveReceivedGiftList{}
	if err = jsonProc(resp, user); err != nil {
		return nil, err
	}
	return user, nil
}
