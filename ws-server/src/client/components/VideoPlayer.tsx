import React, { useEffect, useRef, useState } from 'react';

interface VideoFrame {
  frameData: string; // Base64 encoded image data
  timestamp: string;
}

interface VideoPlayerProps {
  frames: VideoFrame[];
  width?: number;
  height?: number;
  maxFrames?: number; // Maximum number of frames to keep in memory
  autoPlay?: boolean;
  fps?: number;
  showControls?: boolean;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({
  frames,
  width = 800,
  height = 450,
  maxFrames = 100,
  autoPlay = true,
  fps = 10,
  showControls = true,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isPlaying, setIsPlaying] = useState(autoPlay);
  const [currentFrameIndex, setCurrentFrameIndex] = useState(0);
  const animationRef = useRef<number | null>(null);
  const lastFrameTimeRef = useRef<number>(0);
  const frameInterval = 1000 / fps;

  // Update canvas with current frame
  const drawFrame = (frameData: string) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const img = new Image();
    img.onload = () => {
      // Clear canvas
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      
      // Calculate aspect ratio to maintain proportions
      const imgAspectRatio = img.width / img.height;
      const canvasAspectRatio = canvas.width / canvas.height;
      
      let drawWidth, drawHeight, offsetX = 0, offsetY = 0;
      
      if (imgAspectRatio > canvasAspectRatio) {
        // Image is wider than canvas (relative to height)
        drawWidth = canvas.width;
        drawHeight = canvas.width / imgAspectRatio;
        offsetY = (canvas.height - drawHeight) / 2;
      } else {
        // Image is taller than canvas (relative to width)
        drawHeight = canvas.height;
        drawWidth = canvas.height * imgAspectRatio;
        offsetX = (canvas.width - drawWidth) / 2;
      }
      
      // Draw image centered on canvas
      ctx.drawImage(img, offsetX, offsetY, drawWidth, drawHeight);
    };
    img.src = `data:image/jpeg;base64,${frameData}`;
  };

  // Animation loop
  const animate = (timestamp: number) => {
    if (!isPlaying || frames.length === 0) return;
    
    // Calculate time since last frame
    const elapsed = timestamp - lastFrameTimeRef.current;
    
    // If enough time has passed, draw the next frame
    if (elapsed >= frameInterval) {
      // Update the last frame time
      lastFrameTimeRef.current = timestamp - (elapsed % frameInterval);
      
      // Draw the current frame
      if (currentFrameIndex < frames.length) {
        drawFrame(frames[currentFrameIndex].frameData);
        setCurrentFrameIndex((prev) => (prev + 1) % frames.length);
      }
    }
    
    // Continue animation
    animationRef.current = requestAnimationFrame(animate);
  };

  // Start/stop animation based on isPlaying state
  useEffect(() => {
    if (isPlaying) {
      lastFrameTimeRef.current = performance.now();
      animationRef.current = requestAnimationFrame(animate);
    } else if (animationRef.current) {
      cancelAnimationFrame(animationRef.current);
      animationRef.current = null;
    }
    
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
        animationRef.current = null;
      }
    };
  }, [isPlaying, frames]);

  // Draw the latest frame when a new one arrives
  useEffect(() => {
    if (frames.length > 0) {
      // If not playing, just show the latest frame
      if (!isPlaying) {
        drawFrame(frames[frames.length - 1].frameData);
      }
      // If playing, update the current frame index to the latest
      else if (currentFrameIndex >= frames.length || currentFrameIndex === 0) {
        setCurrentFrameIndex(frames.length - 1);
      }
    }
  }, [frames]);

  // Handle play/pause
  const togglePlayPause = () => {
    setIsPlaying(!isPlaying);
  };

  return (
    <div className="video-player">
      <canvas 
        ref={canvasRef} 
        width={width} 
        height={height}
        className="video-canvas"
      />
      
      {showControls && (
        <div className="video-controls">
          <button 
            className={`btn ${isPlaying ? 'btn-danger' : 'btn-success'}`}
            onClick={togglePlayPause}
          >
            {isPlaying ? 'Pause' : 'Play'}
          </button>
          <div className="frame-info">
            Frame: {currentFrameIndex + 1}/{frames.length}
          </div>
        </div>
      )}
    </div>
  );
};

export default VideoPlayer; 