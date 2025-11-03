#ifdef _WIN32

#include <windows.h>
#include <stdio.h>

// Global state for F13 key (VK_F13 = 0x7C)
static int hotkey_pressed = 0;
static HHOOK keyboard_hook = NULL;

// Low-level keyboard hook callback
LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode >= 0) {
        KBDLLHOOKSTRUCT *pKeyboard = (KBDLLHOOKSTRUCT *)lParam;
        
        // Check for F13 key (VK_F13 = 0x7C)
        // You can change this to any other key code
        if (pKeyboard->vkCode == VK_F13) {
            if (wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN) {
                if (!hotkey_pressed) {
                    hotkey_pressed = 1;
                    printf("F13 key pressed\n");
                }
            } else if (wParam == WM_KEYUP || wParam == WM_SYSKEYUP) {
                if (hotkey_pressed) {
                    hotkey_pressed = 0;
                    printf("F13 key released\n");
                }
            }
        }
    }
    
    return CallNextHookEx(keyboard_hook, nCode, wParam, lParam);
}

// Start the keyboard hook
void start_windows_hotkey_listener() {
    printf("Starting Windows hotkey listener (F13)...\n");
    
    // Install the low-level keyboard hook
    keyboard_hook = SetWindowsHookEx(WH_KEYBOARD_LL, KeyboardProc, NULL, 0);
    
    if (keyboard_hook == NULL) {
        fprintf(stderr, "Failed to install keyboard hook. Error: %lu\n", GetLastError());
        return;
    }
    
    printf("Keyboard hook installed successfully\n");
    
    // Message loop to keep the hook active
    MSG msg;
    while (GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
    
    // Cleanup
    if (keyboard_hook != NULL) {
        UnhookWindowsHookEx(keyboard_hook);
    }
}

// Open Windows microphone privacy settings
void open_windows_microphone_settings() {
    printf("Opening Windows microphone settings...\n");
    
    // Open Windows 10/11 microphone privacy settings
    ShellExecute(NULL, "open", "ms-settings:privacy-microphone", NULL, NULL, SW_SHOWNORMAL);
}

// Get the current state of the hotkey
int get_hotkey_state() {
    return hotkey_pressed;
}

#endif // _WIN32
