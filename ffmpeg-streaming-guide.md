# Efficient Cross-Platform Video Streaming with FFmpeg

This guide provides strategies for efficient video streaming using FFmpeg across different platforms.

## 1. Hardware Acceleration

One of the most significant performance improvements comes from using hardware acceleration:

### Hardware Acceleration Options by Platform

1. **NVIDIA GPUs (Windows, Linux, macOS)**
   - NVENC for encoding
   - NVDEC/CUVID for decoding
   - Example: `ffmpeg -hwaccel cuda -i input.mp4 -c:v h264_nvenc output.mp4`

2. **AMD GPUs**
   - AMF on Windows
   - VAAPI/VDPAU on Linux
   - Example: `ffmpeg -hwaccel amf -i input.mp4 -c:v h264_amf output.mp4`

3. **Intel GPUs**
   - Intel Quick Sync (libmfx)
   - Works on Windows, Linux, and macOS
   - Example: `ffmpeg -hwaccel qsv -i input.mp4 -c:v h264_qsv output.mp4`

4. **macOS**
   - VideoToolbox
   - Example: `ffmpeg -hwaccel videotoolbox -i input.mp4 -c:v h264_videotoolbox output.mp4`

5. **Linux**
   - VAAPI (works with Intel, AMD, and some NVIDIA GPUs)
   - Example: `ffmpeg -hwaccel vaapi -hwaccel_device /dev/dri/renderD128 -i input.mp4 -c:v h264_vaapi output.mp4`

## 2. Efficient Codec Selection

### Recommended Codecs

1. **H.264 (libx264)**
   - Most widely supported codec across all platforms
   - Good balance between quality and file size
   - Example: `ffmpeg -i input.mp4 -c:v libx264 -preset fast -crf 23 -c:a aac -b:a 128k output.mp4`

2. **H.265/HEVC**
   - Better compression than H.264 but more CPU intensive
   - Growing support across platforms
   - Example: `ffmpeg -i input.mp4 -c:v libx265 -preset medium -crf 28 -c:a aac -b:a 128k output.mp4`

3. **AV1**
   - Newest codec with excellent compression
   - Limited hardware support but growing
   - Example: `ffmpeg -i input.mp4 -c:v libaom-av1 -crf 30 -b:v 0 -c:a aac output.mp4`

4. **VP9**
   - Good alternative to H.264 with better compression
   - Well supported in browsers
   - Example: `ffmpeg -i input.mp4 -c:v libvpx-vp9 -b:v 1M -c:a libopus output.webm`

### Codec Presets for Streaming

For real-time streaming, use these presets:
- `ultrafast` or `superfast` for lowest latency
- `zerolatency` tune for real-time applications
- Example: `ffmpeg -i input.mp4 -c:v libx264 -preset ultrafast -tune zerolatency -c:a aac output.mp4`

## 3. Streaming Protocols

### Low-Latency Streaming Protocols

1. **RTMP (Real-Time Messaging Protocol)**
   - Good for low-latency streaming
   - Well supported across platforms
   - Uses TCP for reliable delivery
   - Example: `ffmpeg -i input.mp4 -c:v libx264 -preset ultrafast -tune zerolatency -c:a aac -f flv rtmp://server/live/stream`

2. **SRT (Secure Reliable Transport)**
   - Modern protocol designed for low-latency streaming
   - Error recovery and encryption built-in
   - Works well over unreliable networks
   - Example: `ffmpeg -i input.mp4 -c:v libx264 -preset ultrafast -c:a aac -f mpegts srt://server:port?latency=200`

3. **WebRTC**
   - Ultra-low latency (sub-second)
   - Native browser support
   - Requires additional setup with FFmpeg
   - Example: Use with a WebRTC server like Janus or mediasoup

4. **RIST (Reliable Internet Stream Transport)**
   - Designed for reliable video contribution
   - Low latency with error correction
   - Example: `ffmpeg -i input.mp4 -c:v libx264 -c:a aac -f mpegts rist://server:port`

5. **TCP Direct**
   - Simple point-to-point streaming
   - Lower latency than some other protocols
   - Example: `ffmpeg -i input.mp4 -c:v libx264 -preset ultrafast -tune zerolatency -c:a aac -f mpegts tcp://server:port`

## 4. Optimizing for Low Latency

### Key Low-Latency Settings

1. **Encoder Presets**
   - Use `-preset ultrafast` or `-preset superfast` for minimal encoding delay
   - Add `-tune zerolatency` to optimize for real-time streaming
   - Example: `ffmpeg -i input -c:v libx264 -preset ultrafast -tune zerolatency output`

2. **GOP Size (Keyframe Interval)**
   - Use `-g 30` or lower to insert more frequent keyframes
   - Lower values allow viewers to join streams faster
   - Example: `ffmpeg -i input -c:v libx264 -g 30 output`

3. **Buffer Settings**
   - Use `-fflags nobuffer` to reduce buffering delay
   - Set `-bufsize` to a lower value for less buffering
   - Example: `ffmpeg -fflags nobuffer -i input -bufsize 1000k output`

4. **Frame Timing**
   - Use `-vf setpts=0` to display frames as soon as possible
   - Add `-flags low_delay` to force low delay mode
   - Example: `ffmpeg -i input -vf setpts=0 -flags low_delay output`

5. **Analysis Duration**
   - Reduce `-probesize` and `-analyzeduration` for faster stream startup
   - Example: `ffmpeg -probesize 32 -analyzeduration 0 -i input output`

6. **Audio Settings**
   - Use low-latency audio codecs like AAC with minimal frame size
   - Example: `ffmpeg -i input -c:a aac -ar 44100 -c:v libx264 output`

## 5. Complete Cross-Platform Solution

### 1. Webcam/Screen Capture Streaming (Low Latency)

This command works on Windows, macOS, and Linux with appropriate input device names:

```bash
# Windows
ffmpeg -f dshow -video_size 1280x720 -framerate 30 -i video="Your Camera Name" -f dshow -i audio="Your Microphone Name" \
-c:v libx264 -preset ultrafast -tune zerolatency -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f flv rtmp://your-streaming-server/live/stream_key

# macOS
ffmpeg -f avfoundation -framerate 30 -video_size 1280x720 -i "0:0" \
-c:v libx264 -preset ultrafast -tune zerolatency -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f flv rtmp://your-streaming-server/live/stream_key

# Linux
ffmpeg -f v4l2 -framerate 30 -video_size 1280x720 -i /dev/video0 -f alsa -i hw:0 \
-c:v libx264 -preset ultrafast -tune zerolatency -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f flv rtmp://your-streaming-server/live/stream_key
```

### 2. Hardware-Accelerated Streaming

Choose the appropriate hardware acceleration based on your GPU:

```bash
# NVIDIA GPU (Windows, Linux)
ffmpeg -f dshow -video_size 1280x720 -framerate 30 -i video="Your Camera Name" -f dshow -i audio="Your Microphone Name" \
-c:v h264_nvenc -preset llhq -tune zerolatency -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f flv rtmp://your-streaming-server/live/stream_key

# AMD GPU (Windows, Linux)
ffmpeg -f dshow -video_size 1280x720 -framerate 30 -i video="Your Camera Name" -f dshow -i audio="Your Microphone Name" \
-c:v h264_amf -quality speed -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f flv rtmp://your-streaming-server/live/stream_key

# Intel GPU (Windows, Linux, macOS)
ffmpeg -f dshow -video_size 1280x720 -framerate 30 -i video="Your Camera Name" -f dshow -i audio="Your Microphone Name" \
-c:v h264_qsv -preset veryfast -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f flv rtmp://your-streaming-server/live/stream_key

# macOS (VideoToolbox)
ffmpeg -f avfoundation -framerate 30 -video_size 1280x720 -i "0:0" \
-c:v h264_videotoolbox -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f flv rtmp://your-streaming-server/live/stream_key
```

### 3. Point-to-Point Low-Latency Streaming

For direct streaming between computers without a streaming server:

```bash
# Sender (TCP for reliability)
ffmpeg -f dshow -video_size 1280x720 -framerate 30 -i video="Your Camera Name" -f dshow -i audio="Your Microphone Name" \
-c:v libx264 -preset ultrafast -tune zerolatency -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f mpegts tcp://receiver-ip:1234

# Receiver
ffplay -fflags nobuffer -flags low_delay -framedrop tcp://receiver-ip:1234?listen
```

### 4. SRT Protocol for Reliable Low-Latency Streaming

SRT is excellent for unreliable networks:

```bash
# Sender
ffmpeg -f dshow -video_size 1280x720 -framerate 30 -i video="Your Camera Name" -f dshow -i audio="Your Microphone Name" \
-c:v libx264 -preset ultrafast -tune zerolatency -profile:v baseline -pix_fmt yuv420p \
-b:v 2500k -maxrate 2500k -bufsize 5000k -g 30 -keyint_min 30 \
-c:a aac -b:a 128k -ar 44100 -f mpegts srt://receiver-ip:1234?pkt_size=1316&mode=caller&latency=200

# Receiver
ffplay -fflags nobuffer srt://receiver-ip:1234?pkt_size=1316&mode=listener&latency=200
```

### 5. Multi-Bitrate Streaming for Adaptive Playback

For streaming to services that support adaptive bitrate:

```bash
ffmpeg -f dshow -video_size 1920x1080 -framerate 30 -i video="Your Camera Name" -f dshow -i audio="Your Microphone Name" \
-filter_complex "[0:v]split=3[v1][v2][v3]; \
[v1]scale=1920x1080[v1out]; \
[v2]scale=1280x720[v2out]; \
[v3]scale=854x480[v3out]" \
-map "[v1out]" -c:v:0 libx264 -preset veryfast -tune zerolatency -b:v:0 5000k -maxrate:v:0 5000k -bufsize:v:0 10000k -g 60 \
-map "[v2out]" -c:v:1 libx264 -preset veryfast -tune zerolatency -b:v:1 3000k -maxrate:v:1 3000k -bufsize:v:1 6000k -g 60 \
-map "[v3out]" -c:v:2 libx264 -preset veryfast -tune zerolatency -b:v:2 1000k -maxrate:v:2 1000k -bufsize:v:2 2000k -g 60 \
-map 0:a -c:a aac -b:a 128k -ar 44100 \
-f tee "[f=flv]rtmp://your-streaming-server/live/stream_key"
```

## Recommendations for Best Results

1. **Test Your Network Bandwidth**: Ensure your upload speed can handle your chosen bitrate. For reliable streaming, use 70-80% of your available upload bandwidth.

2. **Choose the Right Protocol**:
   - RTMP: Best for streaming to services like YouTube, Twitch, Facebook
   - SRT: Best for unreliable networks or long-distance streaming
   - TCP: Best for point-to-point streaming on reliable networks
   - UDP: Lowest latency but may have packet loss

3. **Hardware Acceleration**:
   - Use hardware acceleration when available to reduce CPU usage
   - Test quality differences between software and hardware encoding

4. **Monitor CPU Usage**:
   - If CPU usage is too high, lower the resolution or framerate
   - Consider using a faster preset at the cost of some quality

5. **Cross-Platform Considerations**:
   - Use platform-specific input devices (dshow on Windows, avfoundation on macOS, v4l2 on Linux)
   - Ensure your FFmpeg build includes support for your chosen hardware acceleration

6. **Playback Compatibility**:
   - Use baseline profile for maximum device compatibility
   - Use yuv420p pixel format for best compatibility 