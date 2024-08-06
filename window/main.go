package window

import (
	"gui/element"
	"gui/fps"
	"hash/fnv"
	"image"
	ic "image/color"
	"math"
	"slices"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Rect struct {
	Node  rl.Rectangle
	Color rl.Color // Added a Color field
	Text  Text
}

type Text struct {
	Color rl.Color
	Size  float32
	Value string
	Font  string
}

// WindowManager manages the window and rectangles
type WindowManager struct {
	Fonts          map[string]rl.Font
	FPSCounterOn   bool
	FPS            int32
	FPSCounter     fps.FPSCounter
	Textures       map[int]Texture
	CanvasTextures map[int]CanvasTexture
}

type Texture struct {
	Hash  uint64
	Image rl.Texture2D
}

type CanvasTexture struct {
	Hash  uint64
	Image rl.Texture2D
}

// NewWindowManager creates a new WindowManager instance
func NewWindowManager() *WindowManager {
	fpsCounter := fps.NewFPSCounter()

	return &WindowManager{
		Fonts:      make(map[string]rl.Font),
		FPSCounter: *fpsCounter,
	}
}

// OpenWindow opens the window
func (wm *WindowManager) OpenWindow(width, height int32) {
	rl.InitWindow(width, height, "")
	rl.SetTargetFPS(30)
	// Enable window resizing
	rl.SetWindowState(rl.FlagWindowResizable)
}

func (wm *WindowManager) SetFPS(fps int32) {
	wm.FPS = fps
	rl.SetTargetFPS(fps)
}

// CloseWindow closes the window
func (wm *WindowManager) CloseWindow() {
	rl.CloseWindow()
}

func (wm *WindowManager) LoadTextures(nodes []element.State) {
	if wm.Textures == nil {
		wm.Textures = make(map[int]Texture)
	}
	if wm.CanvasTextures == nil {
		wm.CanvasTextures = make(map[int]CanvasTexture)
	}

	for i, node := range nodes {
		if node.Texture != nil {
			hash := computeImageHash(node.Texture)
			currentTexture, exists := wm.Textures[i]
			if !exists || currentTexture.Hash != hash {
				if exists {
					rl.UnloadTexture(currentTexture.Image)
				}
				texture := rl.LoadTextureFromImage(rl.NewImageFromImage(node.Texture))
				wm.Textures[i] = Texture{
					Hash:  hash,
					Image: texture,
				}
			}
		}

		if node.Canvas != nil && node.Canvas.Context != nil {
			hash := computeImageHash(node.Canvas.Context)
			currentCanvasTexture, exists := wm.CanvasTextures[i]
			if !exists || currentCanvasTexture.Hash != hash {
				if exists {
					rl.UnloadTexture(currentCanvasTexture.Image)
				}
				texture := rl.LoadTextureFromImage(rl.NewImageFromImage(node.Canvas.Context))
				wm.CanvasTextures[i] = CanvasTexture{
					Hash:  hash,
					Image: texture,
				}
			}
		}
	}
}

// Draw draws all nodes on the window
func (wm *WindowManager) Draw(nodes []element.State) {
	indexes := []float32{0}
	// !TODO: Only Draw whats in fov

	for a := 0; a < len(indexes); a++ {
		for i, node := range nodes {

			if node.Z == indexes[a] {
				p := node.Padding

				DrawRoundedRect(node.X,
					node.Y,
					node.Width+node.Border.Left.Width+node.Border.Right.Width,
					node.Height+node.Border.Top.Width+node.Border.Bottom.Width,
					node.Border.Radius.TopLeft, node.Border.Radius.TopRight, node.Border.Radius.BottomLeft, node.Border.Radius.BottomRight, node.Background)

				// Draw the border based on the style for each side

				if node.Canvas != nil {
					rl.DrawTexture(wm.CanvasTextures[i].Image, int32(node.X), int32(node.Y), rl.White)
				}

				if node.Texture != nil {
					r, g, b, a := node.Color.RGBA()
					rl.DrawTexture(wm.Textures[i].Image, int32(node.X+p.Left+node.Border.Left.Width), int32(node.Y+p.Top+node.Border.Top.Width), ic.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
				}
			} else {
				if !slices.Contains(indexes, node.Z) {
					indexes = append(indexes, node.Z)
					sort.Slice(indexes, func(i, j int) bool {
						return indexes[i] < indexes[j]
					})
				}
			}
		}
	}

	if wm.FPSCounterOn {
		wm.FPSCounter.Update()
		wm.FPSCounter.Draw(10, 10, 10, rl.DarkGray)
	}

}

// WindowShouldClose returns true if the window should close
func (wm *WindowManager) WindowShouldClose() bool {
	return rl.WindowShouldClose()
}

// computeImageHash calculates a hash of the image data
func computeImageHash(img *image.RGBA) uint64 {
	percentage := 20.0
	if percentage <= 0 || percentage > 100 {
		percentage = 100 // Ensure valid percentage range
	}

	totalPixels := len(img.Pix) / 4 // Each pixel consists of 4 bytes (RGBA)

	if totalPixels == 0 {
		return 0
	} else {
		hasher := fnv.New64a()
		step := int(math.Max(1, float64(totalPixels)/(float64(totalPixels)*(percentage/100))))

		// Process a subset of the image data based on the percentage
		for i := 0; i < len(img.Pix); i += step * 4 {
			hasher.Write(img.Pix[i : i+4])
		}

		return hasher.Sum64()
	}

}

func DrawRoundedRect(x, y, width, height float32, topLeftRadius, topRightRadius, bottomLeftRadius, bottomRightRadius float32, color rl.Color) {
	if topLeftRadius > width-bottomLeftRadius {
		topLeftRadius = width - bottomLeftRadius
	}
	if topLeftRadius > height-bottomLeftRadius {
		topLeftRadius = height - bottomLeftRadius
	}

	if topRightRadius > width-bottomRightRadius {
		topRightRadius = width - bottomRightRadius
	}
	if topRightRadius > height-bottomRightRadius {
		topRightRadius = height - bottomRightRadius
	}

	if bottomLeftRadius > width-topLeftRadius {
		bottomLeftRadius = width - topLeftRadius
	}
	if bottomLeftRadius > height-topLeftRadius {
		bottomLeftRadius = height - topLeftRadius
	}

	if bottomRightRadius > width-topRightRadius {
		bottomRightRadius = width - topRightRadius
	}
	if bottomRightRadius > height-topRightRadius {
		bottomRightRadius = height - topRightRadius
	}

	// if topLeftRadius+topRightRadius > width {
	// 	topLeftRadius = width / 2
	// 	topRightRadius = width / 2
	// }
	if bottomLeftRadius+bottomRightRadius > width {
		bottomLeftRadius = width / 2
		bottomRightRadius = width / 2
	}

	if topLeftRadius+bottomLeftRadius > height {
		topLeftRadius = height / 2
		bottomLeftRadius = height / 2
	}
	if topRightRadius+bottomRightRadius > height {
		topRightRadius = height / 2
		bottomRightRadius = height / 2
	}

	// Draw the main rectangle excluding corners
	rl.DrawRectangle(int32(x+topLeftRadius), int32(y), int32(width-topLeftRadius-topRightRadius), int32(height), color)
	rl.DrawRectangle(int32(x), int32(y+topLeftRadius), int32(topLeftRadius), int32(height-topLeftRadius-bottomLeftRadius), color)
	rl.DrawRectangle(int32(x+width-topRightRadius), int32(y+topRightRadius), int32(topRightRadius), int32(height-topRightRadius-bottomRightRadius), color)
	rl.DrawRectangle(int32(x+bottomLeftRadius), int32(y+height-bottomLeftRadius), int32(width-bottomLeftRadius-bottomRightRadius), int32(bottomLeftRadius), color)

	// Draw the corner circles
	rl.DrawCircleSector(rl.Vector2{X: x + topLeftRadius, Y: y + topLeftRadius}, topLeftRadius, 180, 270, 16, color)
	rl.DrawCircleSector(rl.Vector2{X: x + width - topRightRadius, Y: y + topRightRadius}, topRightRadius, 270, 360, 16, color)
	rl.DrawCircleSector(rl.Vector2{X: x + width - bottomRightRadius, Y: y + height - bottomRightRadius}, bottomRightRadius, 0, 90, 16, color)
	rl.DrawCircleSector(rl.Vector2{X: x + bottomLeftRadius, Y: y + height - bottomLeftRadius}, bottomLeftRadius, 90, 180, 16, color)

	// Draw rectangle parts to fill the gaps
	rl.DrawRectangle(int32(x+topLeftRadius), int32(y), int32(width-topLeftRadius-topRightRadius), int32(topLeftRadius), color)                                     // Top
	rl.DrawRectangle(int32(x), int32(y+topLeftRadius), int32(topLeftRadius), int32(height-topLeftRadius-bottomLeftRadius), color)                                  // Left
	rl.DrawRectangle(int32(x+width-topRightRadius), int32(y+topRightRadius), int32(topRightRadius), int32(height-topRightRadius-bottomRightRadius), color)         // Right
	rl.DrawRectangle(int32(x+bottomLeftRadius), int32(y+height-bottomLeftRadius), int32(width-bottomLeftRadius-bottomRightRadius), int32(bottomLeftRadius), color) // Bottom
}
