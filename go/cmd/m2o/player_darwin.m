#import <AVFoundation/AVFoundation.h>
#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

typedef struct {
    AVPlayer *player;
    AVPlayerItem *item;
} M2OPlayer;

static char *copyError(NSString *message) {
    return strdup([message UTF8String]);
}

void *M2OPlayerStart(const char *streamURL, char **errorMessage) {
    @autoreleasepool {
        // A command-line process has no application object by default.
        // Creating it establishes the macOS main run loop used by AVFoundation.
        [NSApplication sharedApplication];
        NSURL *url = [NSURL URLWithString:[NSString stringWithUTF8String:streamURL]];
        if (url == nil) {
            *errorMessage = copyError(@"invalid stream URL");
            return NULL;
        }

        AVPlayerItem *item = [AVPlayerItem playerItemWithURL:url];
        AVPlayer *player = [AVPlayer playerWithPlayerItem:item];
        if (player == nil) {
            *errorMessage = copyError(@"AVFoundation could not create an audio player");
            return NULL;
        }

        // HLS preparation is asynchronous and does not begin until playback is
        // requested. Waiting here turns silent startup
        // failures (bad manifest, unsupported codec, network error) into a
        // useful terminal error rather than leaving a seemingly running player.
        [player play];
        NSDate *deadline = [NSDate dateWithTimeIntervalSinceNow:10.0];
        while ([deadline timeIntervalSinceNow] > 0) {
            if (item.status == AVPlayerItemStatusReadyToPlay) {
                break;
            }
            if (item.status == AVPlayerItemStatusFailed) {
                *errorMessage = copyError(item.error.localizedDescription ?: @"AVFoundation could not load the stream");
                return NULL;
            }
            [[NSRunLoop currentRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:0.1]];
        }
        if (item.status != AVPlayerItemStatusReadyToPlay) {
            *errorMessage = copyError(@"timed out while AVFoundation prepared the HLS stream");
            return NULL;
        }

        M2OPlayer *handle = calloc(1, sizeof(M2OPlayer));
        handle->player = [player retain];
        handle->item = [item retain];
        return handle;
    }
}

void M2OPlayerStop(void *opaquePlayer) {
    @autoreleasepool {
        M2OPlayer *handle = (M2OPlayer *)opaquePlayer;
        [handle->player pause];
        [handle->player release];
        [handle->item release];
        free(handle);
    }
}
