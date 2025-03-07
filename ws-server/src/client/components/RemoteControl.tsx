import React, { useState, useRef, useEffect } from 'react';

interface RemoteControlProps {
  clientId: string;
  screenWidth?: number;
  screenHeight?: number;
  onSendMouseEvent: (clientId: string, action: string, x: number, y: number, button?: string, double?: boolean, amount?: number) => void;
  onSendKeyboardEvent: (clientId: string, action: string, key: string, keys?: string[], text?: string) => void;
  onRequestScreenSize: (clientId: string) => void;
  onRequestMousePosition: (clientId: string) => void;
}

const RemoteControl: React.FC<RemoteControlProps> = ({
  clientId,
  screenWidth,
  screenHeight,
  onSendMouseEvent,
  onSendKeyboardEvent,
  onRequestScreenSize,
  onRequestMousePosition
}) => {
  const [text, setText] = useState('');
  const [key, setKey] = useState('');
  const [mouseX, setMouseX] = useState(0);
  const [mouseY, setMouseY] = useState(0);
  const [scrollAmount, setScrollAmount] = useState(5);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isMouseDown, setIsMouseDown] = useState(false);
  const [selectedButton, setSelectedButton] = useState<'left' | 'right' | 'middle'>('left');
  const [showKeyboard, setShowKeyboard] = useState(false);

  // Common keyboard keys
  const commonKeys = [
    ['escape', 'f1', 'f2', 'f3', 'f4', 'f5', 'f6', 'f7', 'f8', 'f9', 'f10', 'f11', 'f12'],
    ['`', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '-', '=', 'backspace'],
    ['tab', 'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', '[', ']', '\\'],
    ['capslock', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', '\'', 'enter'],
    ['shift', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', 'shift'],
    ['control', 'alt', 'command', 'space', 'command', 'alt', 'left', 'up', 'down', 'right']
  ];

  // Key combinations
  const keyCombinations = [
    { label: 'Copy (Ctrl+C)', keys: ['control', 'c'] },
    { label: 'Paste (Ctrl+V)', keys: ['control', 'v'] },
    { label: 'Cut (Ctrl+X)', keys: ['control', 'x'] },
    { label: 'Select All (Ctrl+A)', keys: ['control', 'a'] },
    { label: 'Undo (Ctrl+Z)', keys: ['control', 'z'] },
    { label: 'Redo (Ctrl+Y)', keys: ['control', 'y'] },
    { label: 'Save (Ctrl+S)', keys: ['control', 's'] },
    { label: 'Print (Ctrl+P)', keys: ['control', 'p'] },
    { label: 'Find (Ctrl+F)', keys: ['control', 'f'] },
    { label: 'Alt+Tab', keys: ['alt', 'tab'] },
    { label: 'Win+D (Show Desktop)', keys: ['command', 'd'] },
    { label: 'Alt+F4 (Close)', keys: ['alt', 'f4'] }
  ];

  // Request screen size when component mounts
  useEffect(() => {
    onRequestScreenSize(clientId);
  }, [clientId, onRequestScreenSize]);

  // Handle canvas click
  const handleCanvasClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!canvasRef.current || !screenWidth || !screenHeight) return;
    
    const canvas = canvasRef.current;
    const rect = canvas.getBoundingClientRect();
    
    // Calculate position relative to canvas
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    
    // Calculate position relative to client's screen
    const targetX = Math.round((x / canvas.width) * screenWidth);
    const targetY = Math.round((y / canvas.height) * screenHeight);
    
    // Send mouse click event
    onSendMouseEvent(clientId, 'click', targetX, targetY, selectedButton, false);
  };

  // Handle canvas mouse down
  const handleCanvasMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!canvasRef.current || !screenWidth || !screenHeight) return;
    
    setIsMouseDown(true);
    
    const canvas = canvasRef.current;
    const rect = canvas.getBoundingClientRect();
    
    // Calculate position relative to canvas
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    
    // Calculate position relative to client's screen
    const targetX = Math.round((x / canvas.width) * screenWidth);
    const targetY = Math.round((y / canvas.height) * screenHeight);
    
    // Send mouse down event
    onSendMouseEvent(clientId, 'down', targetX, targetY, selectedButton, false);
  };

  // Handle canvas mouse move
  const handleCanvasMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!canvasRef.current || !screenWidth || !screenHeight || !isMouseDown) return;
    
    const canvas = canvasRef.current;
    const rect = canvas.getBoundingClientRect();
    
    // Calculate position relative to canvas
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    
    // Calculate position relative to client's screen
    const targetX = Math.round((x / canvas.width) * screenWidth);
    const targetY = Math.round((y / canvas.height) * screenHeight);
    
    // Send mouse move event
    onSendMouseEvent(clientId, 'move', targetX, targetY);
  };

  // Handle canvas mouse up
  const handleCanvasMouseUp = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!canvasRef.current || !screenWidth || !screenHeight || !isMouseDown) return;
    
    setIsMouseDown(false);
    
    const canvas = canvasRef.current;
    const rect = canvas.getBoundingClientRect();
    
    // Calculate position relative to canvas
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    
    // Calculate position relative to client's screen
    const targetX = Math.round((x / canvas.width) * screenWidth);
    const targetY = Math.round((y / canvas.height) * screenHeight);
    
    // Send mouse up event
    onSendMouseEvent(clientId, 'up', targetX, targetY, selectedButton, false);
  };

  // Handle key press
  const handleKeyPress = (keyValue: string) => {
    onSendKeyboardEvent(clientId, 'press', keyValue);
  };

  // Handle key combination
  const handleKeyCombination = (keys: string[]) => {
    if (keys.length < 2) return;
    
    onSendKeyboardEvent(clientId, 'combination', keys[keys.length - 1], keys);
  };

  // Handle text input
  const handleSendText = () => {
    if (!text) return;
    
    onSendKeyboardEvent(clientId, 'type', '', [], text);
    setText('');
  };

  // Handle custom key press
  const handleSendKey = () => {
    if (!key) return;
    
    onSendKeyboardEvent(clientId, 'press', key);
    setKey('');
  };

  // Handle scroll
  const handleScroll = (direction: 'up' | 'down') => {
    const amount = direction === 'up' ? scrollAmount : -scrollAmount;
    onSendMouseEvent(clientId, 'scroll', 0, 0, 'left', false, amount);
  };

  return (
    <div className="remote-control">
      <div className="remote-control-header">
        <h3>Remote Control</h3>
        <div className="remote-control-tabs">
          <button 
            className={`btn btn-sm ${!showKeyboard ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setShowKeyboard(false)}
          >
            Mouse
          </button>
          <button 
            className={`btn btn-sm ${showKeyboard ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setShowKeyboard(true)}
          >
            Keyboard
          </button>
        </div>
      </div>
      
      {!showKeyboard ? (
        <div className="mouse-control">
          <div className="screen-display">
            <canvas 
              ref={canvasRef}
              width={320}
              height={180}
              className="screen-canvas"
              onClick={handleCanvasClick}
              onMouseDown={handleCanvasMouseDown}
              onMouseMove={handleCanvasMouseMove}
              onMouseUp={handleCanvasMouseUp}
              onMouseLeave={handleCanvasMouseUp}
            />
            <div className="screen-overlay">
              {!screenWidth || !screenHeight ? (
                <div className="screen-message">Click to request screen size</div>
              ) : (
                <div className="screen-message">Click to control mouse ({screenWidth}×{screenHeight})</div>
              )}
            </div>
          </div>
          
          <div className="mouse-buttons">
            <div className="button-group">
              <button 
                className={`btn ${selectedButton === 'left' ? 'btn-primary' : 'btn-secondary'} btn-sm`}
                onClick={() => setSelectedButton('left')}
              >
                Left
              </button>
              <button 
                className={`btn ${selectedButton === 'middle' ? 'btn-primary' : 'btn-secondary'} btn-sm`}
                onClick={() => setSelectedButton('middle')}
              >
                Middle
              </button>
              <button 
                className={`btn ${selectedButton === 'right' ? 'btn-primary' : 'btn-secondary'} btn-sm`}
                onClick={() => setSelectedButton('right')}
              >
                Right
              </button>
            </div>
            
            <div className="button-group">
              <button 
                className="btn btn-secondary btn-sm"
                onClick={() => onSendMouseEvent(clientId, 'click', 0, 0, selectedButton, true)}
              >
                Double Click
              </button>
            </div>
          </div>
          
          <div className="scroll-control">
            <div className="scroll-label">Scroll:</div>
            <button 
              className="btn btn-secondary btn-sm"
              onClick={() => handleScroll('up')}
            >
              Up
            </button>
            <input 
              type="number" 
              min="1" 
              max="20" 
              value={scrollAmount} 
              onChange={(e) => setScrollAmount(parseInt(e.target.value) || 5)}
              className="scroll-amount"
            />
            <button 
              className="btn btn-secondary btn-sm"
              onClick={() => handleScroll('down')}
            >
              Down
            </button>
          </div>
          
          <div className="position-control">
            <div className="position-inputs">
              <div className="input-group">
                <label>X:</label>
                <input 
                  type="number" 
                  value={mouseX} 
                  onChange={(e) => setMouseX(parseInt(e.target.value) || 0)}
                />
              </div>
              <div className="input-group">
                <label>Y:</label>
                <input 
                  type="number" 
                  value={mouseY} 
                  onChange={(e) => setMouseY(parseInt(e.target.value) || 0)}
                />
              </div>
            </div>
            <button 
              className="btn btn-primary btn-sm"
              onClick={() => onSendMouseEvent(clientId, 'move', mouseX, mouseY)}
            >
              Move
            </button>
          </div>
          
          <div className="action-buttons">
            <button 
              className="btn btn-secondary btn-sm"
              onClick={() => onRequestScreenSize(clientId)}
            >
              Get Screen Size
            </button>
            <button 
              className="btn btn-secondary btn-sm"
              onClick={() => onRequestMousePosition(clientId)}
            >
              Get Mouse Position
            </button>
          </div>
        </div>
      ) : (
        <div className="keyboard-control">
          <div className="text-input">
            <input 
              type="text" 
              value={text} 
              onChange={(e) => setText(e.target.value)}
              placeholder="Type text to send..."
              className="text-field"
            />
            <button 
              className="btn btn-primary btn-sm"
              onClick={handleSendText}
            >
              Send Text
            </button>
          </div>
          
          <div className="key-input">
            <input 
              type="text" 
              value={key} 
              onChange={(e) => setKey(e.target.value)}
              placeholder="Custom key (e.g., 'escape', 'f1')"
              className="key-field"
            />
            <button 
              className="btn btn-primary btn-sm"
              onClick={handleSendKey}
            >
              Send Key
            </button>
          </div>
          
          <div className="virtual-keyboard">
            {commonKeys.map((row, rowIndex) => (
              <div key={`row-${rowIndex}`} className="key-row">
                {row.map((keyValue, keyIndex) => (
                  <button 
                    key={`key-${rowIndex}-${keyIndex}`}
                    className={`keyboard-key key-${keyValue}`}
                    onClick={() => handleKeyPress(keyValue)}
                  >
                    {keyValue}
                  </button>
                ))}
              </div>
            ))}
          </div>
          
          <div className="key-combinations">
            <h4>Key Combinations</h4>
            <div className="combinations-grid">
              {keyCombinations.map((combo, index) => (
                <button 
                  key={`combo-${index}`}
                  className="btn btn-secondary btn-sm"
                  onClick={() => handleKeyCombination(combo.keys)}
                >
                  {combo.label}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default RemoteControl; 