/*
Everything happening here is by the Grace of God.
*/


#include <stdio.h>
#include <pthread.h>
#include <ApplicationServices/ApplicationServices.h>

int fn_pressed = 0;

int check_accessibility_permission() {
    Boolean trusted = AXIsProcessTrusted();
    if (trusted) {
        printf("Accessibility permission: Authorized\n");
        return 1; 
    } else {
        printf("Accessibility permission: Not authorized\n");
        return 0; 
    }
}

void request_accessibility_permission() {
    printf("Opening accessibility settings...\n");
    
    // Always check first to see current status
    Boolean currentlyTrusted = AXIsProcessTrusted();
    printf("Current trust status: %s\n", currentlyTrusted ? "trusted" : "not trusted");
    
    if (!currentlyTrusted) {
        // Create options dictionary to show prompt
        const void* keys[] = { kAXTrustedCheckOptionPrompt };
        const void* values[] = { kCFBooleanTrue };
        
        CFDictionaryRef options = CFDictionaryCreate(
            kCFAllocatorDefault,
            keys,
            values,
            1,
            &kCFTypeDictionaryKeyCallBacks,
            &kCFTypeDictionaryValueCallBacks
        );
        
        // This will show the system dialog to grant accessibility permissions
        Boolean trustedAfterPrompt = AXIsProcessTrustedWithOptions(options);
        CFRelease(options);
        
        printf("Trust status after prompt: %s\n", trustedAfterPrompt ? "trusted" : "not trusted");
        
        if (!trustedAfterPrompt) {
            printf("User needs to grant accessibility permissions in System Settings\n");
            printf("Opening System Settings directly...\n");
            
            // Also open System Settings to help the user
            CFStringRef urlString = CFSTR("x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility");
            CFURLRef url = CFURLCreateWithString(kCFAllocatorDefault, urlString, NULL);
            
            if (url) {
                LSOpenCFURLRef(url, NULL);
                CFRelease(url);
            }
        }
    } else {
        printf("App is already trusted for accessibility\n");
    }
}

// Function to open accessibility settings directly
void open_accessibility_settings() {
    printf("Opening accessibility settings in System Preferences...\n");
    
    // Open Security & Privacy preferences, specifically the Privacy tab for Accessibility
    CFStringRef urlString = CFSTR("x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility");
    CFURLRef url = CFURLCreateWithString(kCFAllocatorDefault, urlString, NULL);
    
    if (url) {
        OSStatus result = LSOpenCFURLRef(url, NULL);
        if (result == noErr) {
            printf("Successfully opened accessibility settings\n");
        } else {
            printf("Failed to open accessibility settings\n");
        }
        CFRelease(url);
    }
}

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

    printf("Running FN KEY LISTENER..\n");

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

    printf("EVENT TAP CREATED SUCCESSFULLY...\n");

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);

    run_loop(NULL);

    pthread_t thread;
    pthread_create(&thread, NULL, run_loop, NULL);

    return 0;
}