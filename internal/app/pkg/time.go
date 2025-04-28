package pkg

import "time"

func GetCurrentTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Now().In(loc)
}
