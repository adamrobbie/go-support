package screenshot

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kbinani/screenshot"
	"golang.org/x/image/draw"
)

// Quality represents the quality of the screenshot
type Quality string

const (
	// Low quality screenshot (faster, smaller file size)
	Low Quality = "low"
	// Medium quality screenshot
	Medium Quality = "medium"
	// High quality screenshot (slower, larger file size)
	High Quality = "high"
)

// Screenshot represents a captured screenshot
type Screenshot struct {
	Data      []byte    // Raw image data
	Timestamp time.Time // When the screenshot was taken
	Width     int       // Width of the screenshot
	Height    int       // Height of the screenshot
	Format    string    // Format of the screenshot (e.g., "png")
	Quality   Quality   // Quality of the screenshot
}

// Region represents a rectangular region of the screen
type Region struct {
	X      int // X coordinate of the top-left corner
	Y      int // Y coordinate of the top-left corner
	Width  int // Width of the region
	Height int // Height of the region
}

// Capture captures a screenshot with the specified quality
func Capture(quality Quality) (*Screenshot, error) {
	switch runtime.GOOS {
	case "darwin":
		return captureMacOS(quality)
	case "windows":
		return captureWindows(quality)
	case "linux":
		return captureLinux(quality)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// CaptureRegion captures a screenshot of a specific region with the specified quality
func CaptureRegion(region Region, quality Quality) (*Screenshot, error) {
	switch runtime.GOOS {
	case "darwin":
		return captureMacOSRegion(region, quality)
	case "windows":
		return captureWindowsRegion(region, quality)
	case "linux":
		return captureLinuxRegion(region, quality)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// CaptureScreen captures a screenshot of the entire primary display
func CaptureScreen() (image.Image, error) {
	// Get the number of active displays
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return nil, fmt.Errorf("no active displays found")
	}

	// Capture the primary display (index 0)
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return img, nil
}

// CaptureScreenRegion captures a screenshot of a specific region
func CaptureScreenRegion(x, y, width, height int) (image.Image, error) {
	bounds := image.Rect(x, y, x+width, y+height)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("failed to capture region: %w", err)
	}

	return img, nil
}

// captureMacOS captures a screenshot on macOS
func captureMacOS(quality Quality) (*Screenshot, error) {
	// Create a temporary file to store the screenshot
	tmpFile, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Determine the quality settings
	var args []string
	switch quality {
	case Low:
		args = []string{"-x", "-t", "png", "-m", tmpFile.Name()}
	case Medium:
		args = []string{"-x", "-t", "png", tmpFile.Name()}
	case High:
		args = []string{"-x", "-t", "png", "-r", tmpFile.Name()}
	default:
		args = []string{"-x", "-t", "png", tmpFile.Name()}
	}

	// Capture the screenshot
	cmd := exec.Command("screencapture", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Read the screenshot data
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot data: %w", err)
	}

	// Get the image dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &Screenshot{
		Data:      data,
		Timestamp: time.Now(),
		Width:     img.Width,
		Height:    img.Height,
		Format:    "png",
		Quality:   quality,
	}, nil
}

// captureMacOSRegion captures a screenshot of a specific region on macOS
func captureMacOSRegion(region Region, quality Quality) (*Screenshot, error) {
	// Create a temporary file to store the screenshot
	tmpFile, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Determine the quality settings
	var qualityArg string
	switch quality {
	case Low:
		qualityArg = "-m"
	case High:
		qualityArg = "-r"
	default:
		qualityArg = ""
	}

	// Build the command arguments
	args := []string{"-x", "-t", "png"}
	if qualityArg != "" {
		args = append(args, qualityArg)
	}
	args = append(args, "-R", fmt.Sprintf("%d,%d,%d,%d", region.X, region.Y, region.Width, region.Height), tmpFile.Name())

	// Capture the screenshot
	cmd := exec.Command("screencapture", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to capture region screenshot: %w", err)
	}

	// Read the screenshot data
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot data: %w", err)
	}

	// Get the image dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &Screenshot{
		Data:      data,
		Timestamp: time.Now(),
		Width:     img.Width,
		Height:    img.Height,
		Format:    "png",
		Quality:   quality,
	}, nil
}

// captureWindows captures a screenshot on Windows
func captureWindows(quality Quality) (*Screenshot, error) {
	// Create a temporary file to store the screenshot
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("screenshot-%d.png", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	// Determine the quality settings
	var args []string
	switch quality {
	case Low:
		args = []string{"/f", tmpFile, "/d", "/o"}
	case Medium:
		args = []string{"/f", tmpFile, "/d"}
	case High:
		args = []string{"/f", tmpFile}
	default:
		args = []string{"/f", tmpFile}
	}

	// Capture the screenshot using the built-in Windows screenshot tool
	cmd := exec.Command("snippingtool", args...)
	if err := cmd.Run(); err != nil {
		// Try alternative method if snippingtool fails
		cmd = exec.Command("powershell", "-command", fmt.Sprintf("Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('%%{PRTSC}'); Start-Sleep -Milliseconds 500; $img = [System.Windows.Forms.Clipboard]::GetImage(); $img.Save('%s')", tmpFile))
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to capture screenshot: %w", err)
		}
	}

	// Read the screenshot data
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot data: %w", err)
	}

	// Get the image dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &Screenshot{
		Data:      data,
		Timestamp: time.Now(),
		Width:     img.Width,
		Height:    img.Height,
		Format:    "png",
		Quality:   quality,
	}, nil
}

// captureWindowsRegion captures a screenshot of a specific region on Windows
func captureWindowsRegion(region Region, quality Quality) (*Screenshot, error) {
	// Create a temporary file to store the screenshot
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("screenshot-%d.png", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	// Use PowerShell to capture a specific region
	script := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		Add-Type -AssemblyName System.Drawing
		
		$bounds = [System.Drawing.Rectangle]::FromLTRB(%d, %d, %d, %d)
		$bitmap = New-Object System.Drawing.Bitmap $bounds.Width, $bounds.Height
		$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
		
		$graphics.CopyFromScreen($bounds.Left, $bounds.Top, 0, 0, $bounds.Size)
		$bitmap.Save("%s")
		
		$graphics.Dispose()
		$bitmap.Dispose()
	`, region.X, region.Y, region.X+region.Width, region.Y+region.Height, tmpFile)

	cmd := exec.Command("powershell", "-command", script)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to capture region screenshot: %w", err)
	}

	// Read the screenshot data
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot data: %w", err)
	}

	// Get the image dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &Screenshot{
		Data:      data,
		Timestamp: time.Now(),
		Width:     img.Width,
		Height:    img.Height,
		Format:    "png",
		Quality:   quality,
	}, nil
}

// captureLinux captures a screenshot on Linux
func captureLinux(quality Quality) (*Screenshot, error) {
	// Create a temporary file to store the screenshot
	tmpFile, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Try different screenshot tools
	var cmd *exec.Cmd
	var success bool

	// Try gnome-screenshot first
	cmd = exec.Command("gnome-screenshot", "-f", tmpFile.Name())
	if err := cmd.Run(); err == nil {
		success = true
	}

	// Try import (ImageMagick) if gnome-screenshot fails
	if !success {
		qualityArg := "90"
		if quality == High {
			qualityArg = "100"
		} else if quality == Low {
			qualityArg = "75"
		}

		cmd = exec.Command("import", "-window", "root", "-quality", qualityArg, tmpFile.Name())
		if err := cmd.Run(); err == nil {
			success = true
		}
	}

	// Try scrot if import fails
	if !success {
		qualityArg := "90"
		if quality == High {
			qualityArg = "100"
		} else if quality == Low {
			qualityArg = "75"
		}

		cmd = exec.Command("scrot", "-q", qualityArg, tmpFile.Name())
		if err := cmd.Run(); err == nil {
			success = true
		}
	}

	if !success {
		return nil, fmt.Errorf("failed to capture screenshot: no supported screenshot tool found")
	}

	// Read the screenshot data
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot data: %w", err)
	}

	// Get the image dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &Screenshot{
		Data:      data,
		Timestamp: time.Now(),
		Width:     img.Width,
		Height:    img.Height,
		Format:    "png",
		Quality:   quality,
	}, nil
}

// captureLinuxRegion captures a screenshot of a specific region on Linux
func captureLinuxRegion(region Region, quality Quality) (*Screenshot, error) {
	// Create a temporary file to store the screenshot
	tmpFile, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Try different screenshot tools
	var cmd *exec.Cmd
	var success bool

	// Try gnome-screenshot first
	cmd = exec.Command("gnome-screenshot", "-a", "-f", tmpFile.Name())
	if err := cmd.Run(); err == nil {
		success = true
	}

	// Try import (ImageMagick) if gnome-screenshot fails
	if !success {
		qualityArg := "90"
		if quality == High {
			qualityArg = "100"
		} else if quality == Low {
			qualityArg = "75"
		}

		cmd = exec.Command("import", "-quality", qualityArg,
			"-crop", fmt.Sprintf("%dx%d+%d+%d", region.Width, region.Height, region.X, region.Y),
			tmpFile.Name())
		if err := cmd.Run(); err == nil {
			success = true
		}
	}

	// Try scrot if import fails
	if !success {
		qualityArg := "90"
		if quality == High {
			qualityArg = "100"
		} else if quality == Low {
			qualityArg = "75"
		}

		cmd = exec.Command("scrot", "-a",
			fmt.Sprintf("%d,%d,%d,%d", region.X, region.Y, region.Width, region.Height),
			"-q", qualityArg, tmpFile.Name())
		if err := cmd.Run(); err == nil {
			success = true
		}
	}

	if !success {
		return nil, fmt.Errorf("failed to capture region screenshot: no supported screenshot tool found")
	}

	// Read the screenshot data
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot data: %w", err)
	}

	// Get the image dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &Screenshot{
		Data:      data,
		Timestamp: time.Now(),
		Width:     img.Width,
		Height:    img.Height,
		Format:    "png",
		Quality:   quality,
	}, nil
}

// ToBase64 converts the screenshot to a base64-encoded string
func (s *Screenshot) ToBase64() string {
	return base64.StdEncoding.EncodeToString(s.Data)
}

// ToBase64DataURL converts the screenshot to a base64-encoded data URL
func (s *Screenshot) ToBase64DataURL() string {
	return fmt.Sprintf("data:image/%s;base64,%s", s.Format, s.ToBase64())
}

// SaveToFile saves the screenshot to a file
func (s *Screenshot) SaveToFile(filePath string) error {
	return os.WriteFile(filePath, s.Data, 0644)
}

// Resize resizes the screenshot to the specified width and height
// This implementation uses a bilinear interpolation algorithm for better quality
func (s *Screenshot) Resize(width, height int) error {
	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(s.Data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Create a new RGBA image with the specified dimensions
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Bilinear interpolation for better quality resizing
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate source position with floating point precision
			srcX := float64(x) * float64(srcWidth) / float64(width)
			srcY := float64(y) * float64(srcHeight) / float64(height)

			// Get the four surrounding pixels
			x0, y0 := int(srcX), int(srcY)
			x1, y1 := x0+1, y0+1

			// Ensure we don't go out of bounds
			if x1 >= srcWidth {
				x1 = srcWidth - 1
			}
			if y1 >= srcHeight {
				y1 = srcHeight - 1
			}

			// Calculate interpolation weights
			wx := srcX - float64(x0)
			wy := srcY - float64(y0)

			// Get the four surrounding pixels
			c00 := img.At(x0+srcBounds.Min.X, y0+srcBounds.Min.Y)
			c01 := img.At(x0+srcBounds.Min.X, y1+srcBounds.Min.Y)
			c10 := img.At(x1+srcBounds.Min.X, y0+srcBounds.Min.Y)
			c11 := img.At(x1+srcBounds.Min.X, y1+srcBounds.Min.Y)

			// Convert to RGBA values
			r00, g00, b00, a00 := c00.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// Bilinear interpolation for each channel
			r := uint8((float64(r00)*(1-wx)*(1-wy) + float64(r10)*wx*(1-wy) + float64(r01)*(1-wx)*wy + float64(r11)*wx*wy) / 257)
			g := uint8((float64(g00)*(1-wx)*(1-wy) + float64(g10)*wx*(1-wy) + float64(g01)*(1-wx)*wy + float64(g11)*wx*wy) / 257)
			b := uint8((float64(b00)*(1-wx)*(1-wy) + float64(b10)*wx*(1-wy) + float64(b01)*(1-wx)*wy + float64(b11)*wx*wy) / 257)
			a := uint8((float64(a00)*(1-wx)*(1-wy) + float64(a10)*wx*(1-wy) + float64(a01)*(1-wx)*wy + float64(a11)*wx*wy) / 257)

			// Set the pixel in the new image
			newImg.Set(x, y, color.RGBA{r, g, b, a})
		}
	}

	// Encode the resized image
	var buf bytes.Buffer
	if err := png.Encode(&buf, newImg); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Update the screenshot data
	s.Data = buf.Bytes()
	s.Width = width
	s.Height = height

	return nil
}

// Compress compresses the screenshot to reduce its size
// The quality parameter should be between 1 and 100, with 100 being the highest quality
func (s *Screenshot) Compress(quality int) error {
	if quality < 1 || quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100")
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(s.Data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// For better compression, convert to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Update the screenshot data
	s.Data = buf.Bytes()
	s.Format = "jpeg"

	return nil
}

// ConvertToFormat converts the screenshot to the specified format
func (s *Screenshot) ConvertToFormat(format string) error {
	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(s.Data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return fmt.Errorf("failed to encode image as PNG: %w", err)
		}
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return fmt.Errorf("failed to encode image as JPEG: %w", err)
		}
		format = "jpeg"
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Update the screenshot data
	s.Data = buf.Bytes()
	s.Format = format

	return nil
}

// ResizeImage resizes an image to the specified dimensions while maintaining aspect ratio
func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If the image is already smaller than the max dimensions, return it as is
	if width <= maxWidth && height <= maxHeight {
		return img
	}

	// Calculate aspect ratio
	ratio := float64(width) / float64(height)

	// Determine new dimensions while maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = maxWidth
		newHeight = int(float64(newWidth) / ratio)
		if newHeight > maxHeight {
			newHeight = maxHeight
			newWidth = int(float64(newHeight) * ratio)
		}
	} else {
		newHeight = maxHeight
		newWidth = int(float64(newHeight) * ratio)
		if newWidth > maxWidth {
			newWidth = maxWidth
			newHeight = int(float64(newWidth) / ratio)
		}
	}

	// Create a new RGBA image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Resize the image using bilinear interpolation
	draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	return dst
}
