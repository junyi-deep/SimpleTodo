//go:build darwin && !ios

package desktop

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
#include <Cocoa/Cocoa.h>
#include <ApplicationServices/ApplicationServices.h>
static void todoLevel(void* p, int mode) {
 NSWindow* w=(NSWindow*)p;
 [w setLevel: mode==2 ? (CGWindowLevelForKey(kCGDesktopWindowLevelKey)+1) : (mode==1 ? NSFloatingWindowLevel : NSNormalWindowLevel)];
 if(mode==2) [w setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorStationary];
 else [w setCollectionBehavior:NSWindowCollectionBehaviorDefault];
}
static int todoMouseDown(){return CGEventSourceButtonState(kCGEventSourceStateCombinedSessionState,kCGMouseButtonLeft);}
static void todoPointer(double* x,double* y) { CGEventRef event=CGEventCreate(NULL); CGPoint p=CGEventGetLocation(event); *x=p.x;*y=p.y; CFRelease(event); }
*/
import "C"
import "unsafe"

func SetLevel(p unsafe.Pointer, mode string) error {
	m := 0
	if mode == "top" {
		m = 1
	}
	if mode == "desktop" {
		m = 2
	}
	C.todoLevel(p, C.int(m))
	return nil
}
func Pointer() (int, int) { var x, y C.double; C.todoPointer(&x, &y); return int(x), int(y) }
func Supported() bool     { return true }

func MouseDown() bool { return C.todoMouseDown() != 0 }
