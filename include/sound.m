#include <Cocoa/Cocoa.h>

void PlayStartRecordingSound() {
    @autoreleasepool {
        NSSound *sound = [NSSound soundNamed:@"Morse"];
        if (sound) {
            [sound play];
        }
    }
}
