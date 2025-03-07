import React, { useRef, useEffect, useState } from 'react';

interface Screenshot {
  id: string;
  timestamp: string;
  imageUrl: string;
  width: number;
  height: number;
}

interface ScreenshotPlayerProps {
  screenshots: Screenshot[];
  fps?: number;
  autoPlay?: boolean;
  width?: number;
  height?: number;
}

const ScreenshotPlayer: React.FC<ScreenshotPlayerProps> = ({
  screenshots,
  fps = 2,
  autoPlay = false,
  width = 640,
  height = 360
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isPlaying, setIsPlaying] = useState(autoPlay);
  const [currentFrame, setCurrentFrame] = useState(0);
  const [loadedImages, setLoadedImages] = useState<HTMLImageElement[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [progress, setProgress] = useState(0);
  const animationRef = useRef<number | null>(null);

  // Load all images when component mounts or screenshots change
  useEffect(() => {
    if (!screenshots || screenshots.length === 0) {
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setProgress(0);
    
    const images: HTMLImageElement[] = [];
    let loadedCount = 0;

    screenshots.forEach((screenshot, index) => {
      const img = new Image();
      img.crossOrigin = 'anonymous';
      
      img.onload = () => {
        loadedCount++;
        setProgress(Math.floor((loadedCount / screenshots.length) * 100));
        
        if (loadedCount === screenshots.length) {
          setLoadedImages(images);
          setIsLoading(false);
          if (autoPlay) {
            setIsPlaying(true);
          }
        }
      };
      
      img.onerror = () => {
        console.error(`Failed to load image: ${screenshot.imageUrl}`);
        loadedCount++;
        setProgress(Math.floor((loadedCount / screenshots.length) * 100));
        
        if (loadedCount === screenshots.length) {
          setLoadedImages(images.filter(Boolean));
          setIsLoading(false);
        }
      };
      
      img.src = screenshot.imageUrl;
      images[index] = img;
    });

    return () => {
      // Clean up any pending image loads
      images.forEach(img => {
        img.onload = null;
        img.onerror = null;
      });
    };
  }, [screenshots, autoPlay]);

  // Animation loop
  useEffect(() => {
    if (!isPlaying || isLoading || loadedImages.length === 0) {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
        animationRef.current = null;
      }
      return;
    }

    let lastFrameTime = 0;
    const frameInterval = 1000 / fps;

    const animate = (timestamp: number) => {
      if (!lastFrameTime) lastFrameTime = timestamp;
      
      const elapsed = timestamp - lastFrameTime;
      
      if (elapsed > frameInterval) {
        lastFrameTime = timestamp - (elapsed % frameInterval);
        
        // Draw the current frame
        const canvas = canvasRef.current;
        if (canvas) {
          const ctx = canvas.getContext('2d');
          if (ctx) {
            const img = loadedImages[currentFrame];
            if (img) {
              // Clear canvas
              ctx.clearRect(0, 0, canvas.width, canvas.height);
              
              // Calculate aspect ratio to maintain proportions
              const aspectRatio = img.width / img.height;
              let drawWidth = canvas.width;
              let drawHeight = canvas.width / aspectRatio;
              
              if (drawHeight > canvas.height) {
                drawHeight = canvas.height;
                drawWidth = canvas.height * aspectRatio;
              }
              
              // Center the image
              const x = (canvas.width - drawWidth) / 2;
              const y = (canvas.height - drawHeight) / 2;
              
              // Draw the image
              ctx.drawImage(img, x, y, drawWidth, drawHeight);
              
              // Draw timestamp
              const timestamp = new Date(screenshots[currentFrame].timestamp).toLocaleString();
              ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
              ctx.fillRect(x, y + drawHeight - 30, drawWidth, 30);
              ctx.fillStyle = 'white';
              ctx.font = '14px Arial';
              ctx.textAlign = 'center';
              ctx.fillText(timestamp, x + drawWidth / 2, y + drawHeight - 10);
            }
          }
        }
        
        // Move to next frame
        setCurrentFrame((prev) => (prev + 1) % loadedImages.length);
      }
      
      animationRef.current = requestAnimationFrame(animate);
    };
    
    animationRef.current = requestAnimationFrame(animate);
    
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
        animationRef.current = null;
      }
    };
  }, [isPlaying, isLoading, loadedImages, currentFrame, fps, screenshots]);

  const togglePlayPause = () => {
    setIsPlaying(!isPlaying);
  };

  const handleSliderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newFrame = parseInt(e.target.value, 10);
    setCurrentFrame(newFrame);
    
    // If paused, update the canvas immediately
    if (!isPlaying && canvasRef.current && loadedImages[newFrame]) {
      const canvas = canvasRef.current;
      const ctx = canvas.getContext('2d');
      if (ctx) {
        const img = loadedImages[newFrame];
        
        // Clear canvas
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        // Calculate aspect ratio
        const aspectRatio = img.width / img.height;
        let drawWidth = canvas.width;
        let drawHeight = canvas.width / aspectRatio;
        
        if (drawHeight > canvas.height) {
          drawHeight = canvas.height;
          drawWidth = canvas.height * aspectRatio;
        }
        
        // Center the image
        const x = (canvas.width - drawWidth) / 2;
        const y = (canvas.height - drawHeight) / 2;
        
        // Draw the image
        ctx.drawImage(img, x, y, drawWidth, drawHeight);
        
        // Draw timestamp
        const timestamp = new Date(screenshots[newFrame].timestamp).toLocaleString();
        ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
        ctx.fillRect(x, y + drawHeight - 30, drawWidth, 30);
        ctx.fillStyle = 'white';
        ctx.font = '14px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(timestamp, x + drawWidth / 2, y + drawHeight - 10);
      }
    }
  };

  if (screenshots.length === 0) {
    return <div className="screenshot-player-empty">No screenshots available</div>;
  }

  return (
    <div className="screenshot-player">
      <canvas 
        ref={canvasRef} 
        width={width} 
        height={height}
        className="screenshot-canvas"
      />
      
      {isLoading ? (
        <div className="loading-overlay">
          <div className="loading-spinner"></div>
          <div className="loading-text">Loading screenshots... {progress}%</div>
        </div>
      ) : (
        <div className="player-controls">
          <button 
            className="btn btn-primary btn-sm"
            onClick={togglePlayPause}
          >
            {isPlaying ? 'Pause' : 'Play'}
          </button>
          
          <input
            type="range"
            min="0"
            max={loadedImages.length - 1}
            value={currentFrame}
            onChange={handleSliderChange}
            className="frame-slider"
          />
          
          <div className="frame-counter">
            Frame {currentFrame + 1} of {loadedImages.length}
          </div>
        </div>
      )}
    </div>
  );
};

export default ScreenshotPlayer; 