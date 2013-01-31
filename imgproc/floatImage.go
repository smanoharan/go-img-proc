// Package imgproc contains Image Processing operations.
// Preferred internal representation is a floating point representation FloatImage.
// Floating point is used to avoid loss of precision in the intermediate stages,
//  when applying multiple operations to the same image.
// Outputting to jpg/gif/png will result in (unavoidable) loss in precision however.

package imgproc

import (
	"image"
	"image/color"
	"math"
)

// FloatImage represents an image consisting 3 independent intensity planes
// (either RGB or YCrCb based on the original colorModel).
// Each intensity plane consists of an array of intensities, 
// each represented as a float32, a number in the range [0,65536).
// Each intensity plane is stored independently (rather than interleaving)
// which is useful for (the cache locality of) operations which operate on one plane at a time.
type FloatImage struct {
	X, Y, Z       []float32 // intensity planes
	Width, Height int       // dimensions
}

// Construct a new FloatImage of the specified dimensions, with all pixels zero'd.
func NewFloatImage(width, height int) *FloatImage {
	area := width * height
	return &FloatImage{
		X:      make([]float32, area),
		Y:      make([]float32, area),
		Z:      make([]float32, area),
		Width:  width,
		Height: height,
	}
}

// convert an image (read by Decode) into a floatImage
func ImageToFloatImage(img image.Image) *FloatImage {
	b := img.Bounds()
	width := b.Max.X - b.Min.X
	height := b.Max.Y - b.Min.Y

	// init new blank image
	res := NewFloatImage(width, height)

	for yi := 0; yi < height; yi++ {
		for xi := 0; xi < width; xi++ {
			i := yi*width + xi
			// r,g,b are alpha-pre-multiplied, so alpha can be ignored.
			// TODO use YCrCb instead?
			r, g, b, _ := img.At(xi+b.Min.X, yi+b.Min.Y).RGBA()
			res.X[i], res.Y[i], res.Z[i] = float32(r), float32(g), float32(b)
		}
	}

	return res
}

const RGBA_MAX_I = uint8(255)
const RGBA_MAX_F = float64(255)
const SCALE_CONST = float64(256) // converting from [0,65536) to [0,256)

func (img *FloatImage) Bounds() image.Rectangle { return image.Rect(0, 0, img.Width, img.Height) }

func (img *FloatImage) ColorModel() color.Model { return color.RGBAModel }

func (img *FloatImage) At(x, y int) color.Color {
	// a fn for converting from float64 to int
	fti := func(v float32) uint8 {
		return uint8(math.Max(math.Min(RGBA_MAX_F, float64(v)/SCALE_CONST), 0))
	}

	i := x + y*img.Width
	return color.RGBA{fti(img.X[i]), fti(img.Y[i]), fti(img.Z[i]), RGBA_MAX_I}
}
