package main

import (
	"fmt"
	"time"
)

func haveARest() {
	time.Sleep(250 * time.Microsecond)
}

func main()  {
	rpc := NewLibraRPC(nil)
	db := NewDataBaseAdapter("root:test@(127.0.0.1:32773)/libra")

	db.migration()

	errCnt := 0

	for  {
		if errCnt > 10 {
			fmt.Printf("Max Retry Times")
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

		r, err := rpc.GetTransactions(dbLatestVersion + 1, limit, false)
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