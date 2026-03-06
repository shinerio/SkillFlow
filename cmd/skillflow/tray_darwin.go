//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

@interface SkillFlowTrayDelegate : NSObject
@end

static NSStatusItem *skillflowStatusItem = nil;
static NSMenu *skillflowStatusMenu = nil;
static SkillFlowTrayDelegate *skillflowTrayDelegate = nil;

@implementation SkillFlowTrayDelegate
- (void)onShow:(id)sender {
	for (NSWindow *window in [NSApp windows]) {
		[window deminiaturize:nil];
		[window makeKeyAndOrderFront:nil];
	}
	[NSApp activateIgnoringOtherApps:YES];
}

- (void)onQuit:(id)sender {
	[NSApp terminate:nil];
}
@end

static void skillflow_setup_tray(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		if (skillflowStatusItem != nil) {
			return;
		}
		skillflowTrayDelegate = [[SkillFlowTrayDelegate alloc] init];
		skillflowStatusMenu = [[NSMenu alloc] initWithTitle:@"SkillFlow"];

		NSMenuItem *showItem = [[NSMenuItem alloc] initWithTitle:@"Show SkillFlow" action:@selector(onShow:) keyEquivalent:@""];
		[showItem setTarget:skillflowTrayDelegate];
		[skillflowStatusMenu addItem:showItem];

		[skillflowStatusMenu addItem:[NSMenuItem separatorItem]];

		NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit SkillFlow" action:@selector(onQuit:) keyEquivalent:@""];
		[quitItem setTarget:skillflowTrayDelegate];
		[skillflowStatusMenu addItem:quitItem];

		skillflowStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
		NSStatusBarButton *button = [skillflowStatusItem button];
		if (button != nil) {
			[button setTitle:@"SF"];
			[button setToolTip:@"SkillFlow"];
		}
		[skillflowStatusItem setMenu:skillflowStatusMenu];
	});
}

static void skillflow_teardown_tray(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		if (skillflowStatusItem != nil) {
			[[NSStatusBar systemStatusBar] removeStatusItem:skillflowStatusItem];
			skillflowStatusItem = nil;
		}
		skillflowStatusMenu = nil;
		skillflowTrayDelegate = nil;
	});
}
*/
import "C"

func setupTray(_ *App) error {
	C.skillflow_setup_tray()
	return nil
}

func teardownTray() {
	C.skillflow_teardown_tray()
}
