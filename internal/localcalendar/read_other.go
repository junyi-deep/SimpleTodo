//go:build (!darwin && !windows) || ios

package localcalendar

import "fmt"

func Read(ids []string, start, end int64, request bool) (Result, error) {
	return Result{}, fmt.Errorf("此平台不支持读取本机 Outlook 日历")
}
