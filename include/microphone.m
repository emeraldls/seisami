/**
Pookie, Ive no idea what's happening, but I know God is working...
*/


#import <AVFoundation/AVFoundation.h>
#import <AppKit/AppKit.h>
#include <stdio.h>
#include <unistd.h>

// Function to check current microphone authorization status
int check_microphone_permission() {
    AVAuthorizationStatus status = [AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio];
    
    switch (status) {
        case AVAuthorizationStatusAuthorized:
            printf("Microphone access: Authorized\n");
            return 1; // Authorized
        case AVAuthorizationStatusDenied:
            printf("Microphone access: Denied\n");
            return 0; // Denied
        case AVAuthorizationStatusRestricted:
            printf("Microphone access: Restricted\n");
            return 0; // Restricted
        case AVAuthorizationStatusNotDetermined:
            printf("Microphone access: Not determined\n");
            return -1; // Not determined
        default:
            printf("Microphone access: Unknown status\n");
            return 0; // Unknown, treat as denied
    }
}

// Global variables to store the result of the permission request
static int permission_request_result = -1;
static int permission_request_completed = 0;

// Function to request microphone permission (async)
void request_microphone_permission() {
    permission_request_completed = 0;
    permission_request_result = -1;
    
    printf("Requesting microphone permission...\n");
    
    [AVCaptureDevice requestAccessForMediaType:AVMediaTypeAudio completionHandler:^(BOOL granted) {
        permission_request_result = granted ? 1 : 0;
        permission_request_completed = 1;
        
        if (granted) {
            printf("Microphone permission granted\n");
        } else {
            printf("Microphone permission denied\n");
        }
    }];
}

// Function to check if the permission request has completed
int is_permission_request_completed() {
    return permission_request_completed;
}

// Function to get the result of the permission request
int get_permission_request_result() {
    return permission_request_result;
}

// Synchronous function that requests permission and waits for result
int request_microphone_permission_sync() {
    int current_status = check_microphone_permission();
    
    // If already authorized, return immediately
    if (current_status == 1) {
        return 1;
    }
    
    // If not determined, request permission
    if (current_status == -1) {
        request_microphone_permission();
        
        // Wait for completion (with timeout)
        int timeout_seconds = 30;
        int wait_count = 0;
        
        while (!is_permission_request_completed() && wait_count < timeout_seconds * 10) {
            usleep(100000); // Sleep for 0.1 seconds
            wait_count++;
        }
        
        if (is_permission_request_completed()) {
            return get_permission_request_result();
        } else {
            printf("Microphone permission request timed out\n");
            return 0;
        }
    }
    
    // If denied or restricted, return 0
    return 0;
}

// Function to open system preferences for microphone settings
void open_microphone_settings() {
    printf("Opening microphone settings...\n");
    
    // Open Security & Privacy preferences, specifically the Privacy tab
    NSString *urlString = @"x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone";
    NSURL *url = [NSURL URLWithString:urlString];
    
    if ([[NSWorkspace sharedWorkspace] openURL:url]) {
        printf("Successfully opened microphone settings\n");
    } else {
        printf("Failed to open microphone settings\n");
    }
}