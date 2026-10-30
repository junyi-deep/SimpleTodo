//go:build windows

package localcalendar

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

// Classic Outlook's supported object model reads its existing local profile.
// No Save, Delete, Send or account-login methods are invoked.
func Read(ids []string, start, end int64, request bool) (Result, error) {
	b := calendarIDsJSON(ids)
	script := `$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$ids=@(ConvertFrom-Json ([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('` + base64.StdEncoding.EncodeToString(b) + `'))))
$from=[DateTimeOffset]::FromUnixTimeMilliseconds(` + strconv.FormatInt(start, 10) + `).LocalDateTime
$until=[DateTimeOffset]::FromUnixTimeMilliseconds(` + strconv.FormatInt(end, 10) + `).LocalDateTime
try {
 $app=New-Object -ComObject Outlook.Application
 $ns=$app.GetNamespace('MAPI')
 $calendars=@(); $events=@()
 foreach($store in $ns.Stores){
  try{$folder=$store.GetDefaultFolder(9)}catch{continue}
  $id=$store.StoreID+':'+$folder.EntryID
  $calendars+=@{id=$id;name=$folder.Name;source=$store.DisplayName}
  if($ids -notcontains $id){continue}
  $items=$folder.Items; $items.Sort('[Start]'); $items.IncludeRecurrences=$true
  $format='g';$filter="[Start] < '"+$until.ToString($format)+"' AND [End] > '"+$from.ToString($format)+"'"
  $subset=$items.Restrict($filter)
  foreach($ev in $subset){
   if($events.Count -ge 10000){throw 'TOO_MANY_EVENTS'}
   if($ev.Class -ne 26){continue}
   $begin=([DateTimeOffset]$ev.Start).ToUnixTimeMilliseconds();$finish=([DateTimeOffset]$ev.End).ToUnixTimeMilliseconds()
   $events+=@{id=($id+':'+$ev.EntryID+':'+$begin);calendarId=$id;title=[string]$ev.Subject;location=[string]$ev.Location;start=$begin;end=$finish;allDay=[bool]$ev.AllDayEvent}
  }
 }
 @{calendars=@($calendars);events=@($events)}|ConvertTo-Json -Depth 5 -Compress
}catch{[Console]::Error.Write('需要已安装并配置账户的经典 Outlook；新版 Outlook 不支持本地读取。');exit 1}
`
	units := utf16.Encode([]rune(script))
	encoded := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(encoded[i*2:], u)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(encoded))
	hideCalendarProcess(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Result{}, fmt.Errorf("需要已安装并配置账户的经典 Outlook；新版 Outlook 不支持本地读取")
	}
	if stdout.Len() > 16*1024*1024 {
		return Result{}, fmt.Errorf("日程数据过多")
	}
	var result Result
	err := json.Unmarshal([]byte(strings.TrimPrefix(stdout.String(), "\ufeff")), &result)
	return result, err
}
