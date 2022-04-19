# LiveKit Command Line Client
This client joins a LiveKit room and decodes audio through gstreamer

Based on [Pion gstreamer-receive](https://github.com/pion/example-webrtc-applications/tree/master/gstreamer-receive)

## Dependencies
* Debian - `sudo apt-get install libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev gstreamer1.0-plugins-good`
* macOS - `brew install gst-plugins-good`

## Configuration
Set the following variables in the .env file
```
LIVEKIT_URL=
LIVEKIT_ROOM=
LIVEKIT_ID=
LIVEKIT_API_KEY=
LIVEKIT_API_SECRET=
```

## Debugging
Set gstreamer debug level and set plugin path
`GST_DEBUG=4 GST_PLUGIN_PATH=/opt/homebrew/lib/gstreamer-1.0/`