#import <Cocoa/Cocoa.h>
#import <MediaPlayer/MediaPlayer.h>

#import "nowplaying_darwin.h"
#import "_cgo_export.h"

@interface SpotgoAppDelegate : NSObject <NSApplicationDelegate>
@end

@implementation SpotgoAppDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)notification {
  goAppReady();
}
@end

// Retained for the process lifetime so the delegate isn't deallocated.
static SpotgoAppDelegate *gDelegate = nil;

// The current track's album art, re-applied on every npSetNowPlaying so it
// survives the dictionary being rebuilt each poll. Set asynchronously once the
// image has downloaded.
static MPMediaItemArtwork *gArtwork = nil;

static void npRegisterCommands(void) {
  MPRemoteCommandCenter *center = [MPRemoteCommandCenter sharedCommandCenter];

  center.playCommand.enabled = YES;
  [center.playCommand
      addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goMediaCommand(NP_CMD_PLAY);
        return MPRemoteCommandHandlerStatusSuccess;
      }];

  center.pauseCommand.enabled = YES;
  [center.pauseCommand
      addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goMediaCommand(NP_CMD_PAUSE);
        return MPRemoteCommandHandlerStatusSuccess;
      }];

  center.togglePlayPauseCommand.enabled = YES;
  [center.togglePlayPauseCommand
      addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goMediaCommand(NP_CMD_TOGGLE);
        return MPRemoteCommandHandlerStatusSuccess;
      }];

  center.nextTrackCommand.enabled = YES;
  [center.nextTrackCommand
      addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goMediaCommand(NP_CMD_NEXT);
        return MPRemoteCommandHandlerStatusSuccess;
      }];

  center.previousTrackCommand.enabled = YES;
  [center.previousTrackCommand
      addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goMediaCommand(NP_CMD_PREV);
        return MPRemoteCommandHandlerStatusSuccess;
      }];
}

void npRun(void) {
  @autoreleasepool {
    NSApplication *app = [NSApplication sharedApplication];
    // Accessory: a real app (so Control Center surfaces us) but with no Dock
    // icon and no focus stealing.
    [app setActivationPolicy:NSApplicationActivationPolicyAccessory];
    gDelegate = [[SpotgoAppDelegate alloc] init];
    app.delegate = gDelegate;
    npRegisterCommands();
    [app run];
  }
}

void npStop(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    [NSApp terminate:nil];
  });
}

void npSetNowPlaying(const char *title, const char *artist, const char *album,
                     double durationSec, double elapsedSec, double rate,
                     int state) {
  @autoreleasepool {
    // Build NSStrings now (on the caller's thread); the C pointers are only
    // valid for this call. dispatch_async copies the block, retaining these.
    NSString *nsTitle = title ? @(title) : @"";
    NSString *nsArtist = artist ? @(artist) : @"";
    NSString *nsAlbum = album ? @(album) : @"";
    dispatch_async(dispatch_get_main_queue(), ^{
      NSMutableDictionary *info = [@{
        MPMediaItemPropertyTitle : nsTitle,
        MPMediaItemPropertyArtist : nsArtist,
        MPMediaItemPropertyAlbumTitle : nsAlbum,
        MPMediaItemPropertyPlaybackDuration : @(durationSec),
        MPNowPlayingInfoPropertyElapsedPlaybackTime : @(elapsedSec),
        MPNowPlayingInfoPropertyPlaybackRate : @(rate),
      } mutableCopy];
      if (gArtwork != nil) {
        info[MPMediaItemPropertyArtwork] = gArtwork;
      }
      MPNowPlayingInfoCenter *center = [MPNowPlayingInfoCenter defaultCenter];
      center.nowPlayingInfo = info;
      center.playbackState = (MPNowPlayingPlaybackState)state;
    });
  }
}

void npClearNowPlaying(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    gArtwork = nil;
    MPNowPlayingInfoCenter *center = [MPNowPlayingInfoCenter defaultCenter];
    center.nowPlayingInfo = @{};
    center.playbackState = MPNowPlayingPlaybackStateStopped;
  });
}

void npSetArtwork(const void *data, int len) {
  if (data == NULL || len <= 0) {
    return;
  }
  @autoreleasepool {
    // Copy the bytes now; the C pointer is only valid for this call.
    NSData *imageData = [NSData dataWithBytes:data length:(NSUInteger)len];
    dispatch_async(dispatch_get_main_queue(), ^{
      NSImage *image = [[NSImage alloc] initWithData:imageData];
      if (image == nil) {
        return;
      }
      gArtwork = [[MPMediaItemArtwork alloc]
          initWithBoundsSize:image.size
              requestHandler:^NSImage *(CGSize size) {
                return image;
              }];
      // Merge into the current info so the cover appears without waiting for the
      // next poll.
      MPNowPlayingInfoCenter *center = [MPNowPlayingInfoCenter defaultCenter];
      NSMutableDictionary *info = [center.nowPlayingInfo mutableCopy];
      if (info == nil) {
        return;
      }
      info[MPMediaItemPropertyArtwork] = gArtwork;
      center.nowPlayingInfo = info;
    });
  }
}

void npClearArtwork(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    gArtwork = nil;
    MPNowPlayingInfoCenter *center = [MPNowPlayingInfoCenter defaultCenter];
    NSMutableDictionary *info = [center.nowPlayingInfo mutableCopy];
    if (info != nil) {
      [info removeObjectForKey:MPMediaItemPropertyArtwork];
      center.nowPlayingInfo = info;
    }
  });
}
