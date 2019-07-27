package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"io.librablock.go/controllers"
	"io.librablock.go/utils"
)

func haveARest() {
	time.Sleep(250 * time.Microsecond)
}

func main() {
	botKey := os.Getenv("LIBRA_BOT_KEY")
	botSecret := os.Getenv("LIBRA_BOT_SECRET")
	chatId := os.Getenv("LIBRA_BOT_CHAT_ID")
	dbURL := os.Getenv("LIBRA_MYSQL_URL")

	telegramURL := fmt.Sprintf("https://api.telegram.org/%s:%s/sendMessage?chat_id=%s&parse_mode=markdown&text=", botKey, botSecret, chatId)
	fmt.Println(telegramURL)

	rpc := controllers.NewLibraRPC(nil)
	db := utils.NewDataBaseAdapter(dbURL)

	db.Migration()

	errCnt := 0

	for {
		if errCnt > 10 {
			fmt.Printf("Max Retry Times")

			_, _ = http.Get(telegramURL + url.QueryEscape("libra block fetcher failed 10 times"))
			break
		}

		latestVersion, err := rpc.GetLatestVersion()
		if err != nil {
			errCnt += 1

			haveARest()
			continue
		}

		dbLatestVersion := db.GetLatestVersion()

		limit := latestVersion - dbLatestVersion

		if limit == 0 {
			haveARest()
			errCnt = 0
			continue
		}

		if limit > 1000 {
			limit = 1000
		}

		r, err := rpc.GetTransactions(dbLatestVersion+1, limit, false)
		if err != nil {
			errCnt += 1
			haveARest()
			continue
		}

		for _, v := range *r {
			db.SaveBlock(v)
			fmt.Printf("Success Fetch Version: %d\n", v.Version)
		}
		errCnt = 0
	}

}
