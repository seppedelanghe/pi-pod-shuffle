# Pi Pod Shuffle

## Linux setup
- Install Raspberry OS Lite on SD Card
  - username: `pipod`
  - hostname: `pi-pod-shuffle`
- Connect with SSH

### Install VIM (duh)
```
sudo apt update
sudo apt install vim
```


### Enable bluetooth
```zsh
sudo apt update
sudo apt install bluez-alsa-utils libasound2-plugin-bluez
sudo rfkill unblock bluetooth
```

### Enable BlueALSA:
```zsh
sudo systemctl enable bluealsa
sudo systemctl start bluealsa
sudo systemctl restart bluetooth
```

__confirm BlueALSA is working__:
```zsh
aplay -L | grep bluealsa
# you should see
/> bluealsa
```

### Enable A2DP sink
```zsh
sudo vim /etc/default/bluez-alsa
# set OPTIONS="--profile=a2dp-sink"
sudo systemctl restart bluealsa
```

### Add user to audio group and linger
```
sudo usermod -a -G audio,bluetooth $USER
loginctl enable-linger youruser
```

### Pair headphones
```zsh
bluetoothctl
> agent on
> default-agent
> scan on
# Put headphones in paring mode and look for your headphones
> pair [MAC ADDRESS]
> trust [MAC ADDRESS]
> connect [MAC ADDRESS]
```

### Force audio to BlueALSA
```
sudo vim /etc/asound.conf
```

Paste these contents:
_make sure to replace `XX:XX:XX:XX:XX:XX` with your own headpones their MAC address
```
defaults.bluealsa.interface "hci0"
defaults.bluealsa.device "XX:XX:XX:XX:XX:XX"
defaults.bluealsa.profile "a2dp"
defaults.bluealsa.delay 10000

# Define the BlueALSA PCM
pcm.btheadset {
    type plug
    slave.pcm {
        type bluealsa
        device "XX:XX:XX:XX:XX:XX"
        profile "a2dp"
    }
    hint {
        show on
        description "BlueALSA Audio Device"
    }
}

# Make BlueALSA the system default for this user
pcm.!default {
    type plug
    slave.pcm "btheadset"
}

ctl.!default {
    type bluealsa
}
```

### Build and copy

On your PC:
```zsh
# make sure docker is running
make pi-deploy
```

### Create auto start services
#### Copy files
On your PC:
```
scp ./scripts/* pipod@pi-pod-shuffle.local:/home/pipod/
```

On pi:
```
sudo chmod +x ./*.sh
sudo mv ./*.sh /usr/local/bin/
sudo mv ./*.service /etc/systemd/system/
```

#### activate services
```
sudo systemctl daemon-reload
sudo systemctl enable bt-connect.service
sudo systemctl start bt-connect.service
sudo systemctl enable wifi-ssh-guard.service
sudo systemctl start wifi-ssh-guard.service
```

### Prep music library
On your PC, go to the `desktop` folder.
- Create python venv (`python -m venv .venv`)
- activate venv (source `.venv/bin/activate`)
- install dependencies `pip install -r requirements.txt`
- download the `Cnn6_mAP=0.343.pth` model [source](https://zenodo.org/records/3987831)
- process your music library `python pipod_manager.py --dir <your-music-dir> process`
  - this will create 2 files `raw_features.json` and `library.json`
- Copy your music library (with JSON files) to your pi at `/home/pipod/`
  - For example: `/home/pipod/music/`
  - the `music/` folder includes the `library.json` and all your music files (.mp3, .flac, etc.)
- Check if the `dir` key in the `library.json` file points to the correct path on the pi

#### Reboot and test
```
sudo reboot now
```
The Pi should now reboot, connect to your headphones and start playing music!


## Battery power

To reduce energy consumption of the Pi you can do the following if you are running it of a battery.

### Disable hardware
We do not need HDMI output, USB data and the status LED, so we can disable it all together.
Edit `/boot/firmware/config.txt` and add:
```zsh
# Disable HDMI output
dtoverlay=disable-hdmi

# Disable the USB controller
dtoverlay=dwc2,dr_mode=peripheral

# Disable the Activity LED
dtparam=act_led_trigger=none
dtparam=act_led_activelow=off
```

### CPU downscaling
Once setup us complete we do no longer need all the performance of the Pi Zero 2 W's 4-cores.
Edit `/boot/firmware/config.txt`:
```zsh
arm_freq=800  # Cap clock speed to 800MHz (Default is 1000MHz)
force_turbo=0 # Ensure it scales down when idle
```

__Make it single core:__
Edit `/boot/cmdline.txt` (do not add newlines)
```zsh
maxcpus=1
```

### Enable Read-Only Mode
To prevent SD card corruption during battery depletion:
1. Run `sudo raspi-config`
2. Go to **Performance Options** -> **Overlay File System**
3. Enable the Overlay and set the Boot partition to Read-Only.
*Note: All files saved to the disk will be lost on reboot once this is active.*