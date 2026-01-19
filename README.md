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
sudo apt install pulseaudio pulseaudio-module-bluetooth
sudo rfkill unblock bluetooth
```


### Add user to audio group
```
sudo usermod -a -G audio,bluetooth $USER
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

### Force ALSA to PulseAudio
```
sudo vim /etc/asound.conf
```

Add:
```
pcm.!default {
    type pulse
    fallback "sysdefault"
}

ctl.!default {
    type pulse
}
```
