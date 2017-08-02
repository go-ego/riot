package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/huichen/gobo"
	"github.com/huichen/gobo/contrib"
)

var (
	access_token = flag.String("access_token", "", "用户的访问令牌")
	weibo        = gobo.Weibo{}
	users_file   = flag.String("users_file", "users.txt", "从该文件读入要下载的微博用户名，每个名字一行")
	output_file  = flag.String("output_file", "weibo_data.txt", "将抓取的微博写入下面的文件")
	num_weibos   = flag.Int("num_weibos", 2000, "从每个微博账号中抓取多少条微博")
)

func main() {
	flag.Parse()

	// 读取用户名
	content, err := ioutil.ReadFile(*users_file)
	if err != nil {
		log.Fatal("无法读取-users_file")
	}
	users := strings.Split(string(content), "\n")

	outputFile, _ := os.Create(*output_file)
	defer outputFile.Close()

	// 抓微博
	for _, user := range users {
		if user == "" {
			continue
		}
		log.Printf("抓取 @%s 的微博", user)
		statuses, err := contrib.GetStatuses(
			&weibo, *access_token, user, 0, *num_weibos, 5000) // 超时5秒
		if err != nil {
			log.Print(err)
			continue
		}

		for _, status := range statuses {
			t, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", status.Created_At)
			outputFile.WriteString(fmt.Sprintf(
				"%d||||%d||||%d||||%s||||%d||||%d||||%d||||%s||||%s||||%s\n",
				status.Id, uint32(t.Unix()), status.User.Id, status.User.Screen_Name,
				status.Reposts_Count, status.Comments_Count, status.Attitudes_Count,
				status.Thumbnail_Pic, status.Original_Pic, status.Text))
		}
	}
}
