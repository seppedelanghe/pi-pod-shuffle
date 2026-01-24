# Pi Pod Shuffle


## Linux setup
- Install Raspberry OS Lite on SD Card
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
```
you should see:
```
bluealsa
```

### Enable A2DP sing
- `sudo vim /etc/default/bluez-alsa`
- __SET__: `OPTIONS="--profile=a2dp-sink"`
- `sudo systemctl restart bluealsa`

### Add user to audio group
```
sudo usermod -a -G audio,bluetooth $USER
```

### Force audio to BlueALSA
```
sudo vim ~/.asoundrc
```

Contents:
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

### Connect headphone
```
bluetoothctl
> agent on
> scan on
> pair [MAC ADDRESS]
> trust [MAC ADDRESS]
> connect [MAC ADDRESS]
```


