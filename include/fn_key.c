#include <stdio.h>
#include <pthread.h>
#include <ApplicationServices/ApplicationServices.h>

int fn_pressed = 0;

// Event callback
CGEventRef event_callback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventFlagsChanged) {
        CGEventFlags flags = CGEventGetFlags(event);
        if ((flags & kCGEventFlagMaskSecondaryFn) && !fn_pressed) {
            fn_pressed = 1;
            printf("FN key pressed\n");
        } else if (!(flags & kCGEventFlagMaskSecondaryFn) && fn_pressed) {
            fn_pressed = 0;
            printf("FN key released\n");
        }
    }
    return event;
}

void* run_loop(void* arg) {
    CFRunLoopRun();
    return NULL;
}

int start_listener() {
    // Check for accessibility permissions
    if (!AXIsProcessTrusted()) {
        fprintf(stderr, "Accessibility permissions are required. Please grant them in System Settings.\n");
    }

    printf("Running FN KEY LISTENER\n");

    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        0,
        CGEventMaskBit(kCGEventFlagsChanged),
        event_callback,
        NULL
    );

    if (!eventTap) {
        fprintf(stderr, "Failed to create event tap. This might be due to lack of accessibility permissions.\n");
        return 1;
    }

    printf("EVENT TAP CREATED SUCCESSFULLY\n");

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);

    run_loop(NULL);

    pthread_t thread;
    pthread_create(&thread, NULL, run_loop, NULL);

    return 0;
}