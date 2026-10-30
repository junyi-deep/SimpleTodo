//go:build darwin && !ios

package localcalendar

/*
#cgo CFLAGS: -x objective-c -fblocks
#cgo LDFLAGS: -framework Foundation -framework EventKit
#include <Foundation/Foundation.h>
#include <EventKit/EventKit.h>
#include <stdlib.h>
static char* readLocalCalendar(const char* idsJSON, double start, double end, int request) {
 @autoreleasepool {
  static EKEventStore* store = nil;
  if (!store) store = [[EKEventStore alloc] init];
  EKAuthorizationStatus status = [EKEventStore authorizationStatusForEntityType:EKEntityTypeEvent];
  BOOL allowed = status == EKAuthorizationStatusAuthorized;
  if (@available(macOS 14.0,*)) allowed = status == EKAuthorizationStatusFullAccess;
  if (!allowed && request && (status == EKAuthorizationStatusNotDetermined || status == EKAuthorizationStatusWriteOnly)) {
   dispatch_semaphore_t sem = dispatch_semaphore_create(0);
   __block BOOL granted = NO;
   void (^completion)(BOOL,NSError*) = ^(BOOL yes,NSError* error) { granted=yes;dispatch_semaphore_signal(sem); };
   if (@available(macOS 14.0,*)) [store requestFullAccessToEventsWithCompletion:completion];
   else [store requestAccessToEntityType:EKEntityTypeEvent completion:completion];
   if (dispatch_semaphore_wait(sem,dispatch_time(DISPATCH_TIME_NOW,120*NSEC_PER_SEC)) != 0) return strdup("{\"error\":\"日历授权等待超时，请重试\"}");
   allowed=granted; [store reset];
  }
  if (!allowed) return strdup("{\"error\":\"需要系统日历读取权限，请在系统设置的隐私与安全性中允许日历访问\"}");
  NSArray* ids=[NSJSONSerialization JSONObjectWithData:[[NSString stringWithUTF8String:idsJSON] dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil];
  NSMutableArray* calendars=[NSMutableArray array];NSMutableArray* chosen=[NSMutableArray array];
  for (EKCalendar* cal in [store calendarsForEntityType:EKEntityTypeEvent]) {
   [calendars addObject:@{@"id":cal.calendarIdentifier,@"name":cal.title?:@"日历",@"source":cal.source.title?:@"本机"}];
   if ([ids containsObject:cal.calendarIdentifier]) [chosen addObject:cal];
  }
  NSMutableArray* events=[NSMutableArray array];
  if (chosen.count) {
   NSPredicate* pred=[store predicateForEventsWithStartDate:[NSDate dateWithTimeIntervalSince1970:start] endDate:[NSDate dateWithTimeIntervalSince1970:end] calendars:chosen];
   for (EKEvent* ev in [store eventsMatchingPredicate:pred]) {
    if (events.count>=10000) return strdup("{\"error\":\"日程过多，请缩小读取范围\"}");
    NSString* ident=[NSString stringWithFormat:@"%@:%@:%0.f",ev.calendar.calendarIdentifier,ev.calendarItemIdentifier,ev.startDate.timeIntervalSince1970*1000];
    [events addObject:@{@"id":ident,@"calendarId":ev.calendar.calendarIdentifier,@"title":ev.title?:@"未命名日程",@"location":ev.location?:@"",@"start":@((long long)(ev.startDate.timeIntervalSince1970*1000)),@"end":@((long long)(ev.endDate.timeIntervalSince1970*1000)),@"allDay":@(ev.allDay)}];
   }
  }
  NSData* out=[NSJSONSerialization dataWithJSONObject:@{@"calendars":calendars,@"events":events} options:0 error:nil];
  return strdup([[[NSString alloc] initWithData:out encoding:NSUTF8StringEncoding] autorelease].UTF8String);
 }
}
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"
)

var nativeMu sync.Mutex

func Read(ids []string, start, end int64, request bool) (Result, error) {
	nativeMu.Lock()
	defer nativeMu.Unlock()
	b := calendarIDsJSON(ids)
	input := C.CString(string(b))
	defer C.free(unsafe.Pointer(input))
	req := 0
	if request {
		req = 1
	}
	out := C.readLocalCalendar(input, C.double(start)/1000, C.double(end)/1000, C.int(req))
	defer C.free(unsafe.Pointer(out))
	var result Result
	if err := json.Unmarshal([]byte(C.GoString(out)), &result); err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, fmt.Errorf("%s", result.Error)
	}
	return result, nil
}
