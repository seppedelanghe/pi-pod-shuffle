# Pi Pod Shuffle

## Installation guide
You can find the setup in this [Readme](pi/README.md)

**Pi Pod Shuffle** is a screenless Raspberry Pi music player inspired by the iPod Shuffle.

It’s a simple brick with:
- USB-C power
- On/off switch
- No buttons
- No screen

It connects to Bluetooth headphones and immediately starts playing music.  
Playback, skipping, and volume are controlled entirely via your headphones.

## Smart Shuffle

Tracks how long you listen before skipping:
- **Instant skip** → wrong vibe, switch genres
- **Short listen** → close, adjust slightly
- **Long listen** → good match, play similar tracks

Songs are scored dynamically to match your current mood, producing a smart, adaptive shuffle instead of random playback.

# Sync software

This is very much a work in progress, but the goal is to build some CLI (or GUI) tool to help with syncing music and embeddings to the pi.
There is a lot of prototype code in the `desktop` folder. Generating embeddings does work as I use it for my pi pod shuffle but I transfer files manually. (eg. rsync)

If you do not want to fiddle with the embeddings and AI models to extract them, you can update the code in `pi` before building so that it uses regular shuffle. That removes the need for the embeddings file.


## Disclaimer

This project is provided **as-is**, **without any warranty**, express or implied.

By using this code, you acknowledge and agree that:

- **No support is provided.**  
  There is no guarantee of help, troubleshooting, updates, or maintenance.

- **No warranty of any kind.**  
  This includes (but is not limited to) warranties of merchantability, fitness for a particular purpose, or non-infringement.

- **You assume all risk.**  
  Running this software may cause unexpected behavior, data loss, hardware damage, SD card corruption, or render your Raspberry Pi or connected devices unusable.

- **You are responsible for your hardware.**  
  If this project bricks your Pi, fries a component, corrupts your filesystem, or causes any other damage — that’s on you.

This project is experimental, hardware-adjacent, and intended for users who are comfortable debugging Linux, embedded systems, and broken setups.

If you are not willing to accept the risk of breaking your own device, **do not use this project**.
