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

const TOLERANCE = float64(0.0000001) // for comparing floating point numbers

// FloatImage represents an image consisting 3 independent intensity planes
// (either RGB or YCrCb based on the original colorModel).
// Each intensity plane consists of an array of intensities, 
// each represented as a float32, a number in the range [0,65536).
// Each intensity plane is stored independently (rather than interleaving)
// which is useful for (the cache locality of) operations which operate on one plane at a time.
type FloatImage struct {
	Ip            [3][]float32 // intensity planes
	Width, Height int          // dimensions
}

// Construct a new FloatImage of the specified dimensions, with all pixels zero'd.
func NewFloatImage(width, height int) *FloatImage {
	area := width * height
	return &FloatImage{
		Ip:     [3][]float32{make([]float32, area), make([]float32, area), make([]float32, area)},
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
			res.Ip[0][i], res.Ip[1][i], res.Ip[2][i] = float32(r), float32(g), float32(b)
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
	return color.RGBA{fti(img.Ip[0][i]), fti(img.Ip[1][i]), fti(img.Ip[2][i]), RGBA_MAX_I}
}

func (img *FloatImage) Clone() *FloatImage {
	res := NewFloatImage(img.Width, img.Height)
	for i := 0; i < 3; i++ {
		copy(res.Ip[i], img.Ip[i]) // NOTE: copy args are (dst, src)
	}
	return res
}

// A ConvKernel is a kernel (a NxN matrix) for a Convolution operation.
// The NxN matrix is stored as a 1D array in row-major order.
// (I.e. index-of(x,y) is (y*WIDTH + x))
// In order to ensure the matrix can be centered on a pixel, 
// the size of the matrix must be odd (i.e. N = 2R + 1, for some Radius R)
// Thus, the matrix has (2*Radius + 1)^2 elements.
// For example, a 3x3 matrix has radius of 1 and has 9 elements.
type ConvKernel struct {
	Kernel []float32
	Radius int
}

// a convenience function for creating 3x3 convolution kernels.
func NewConvKernel3(m11, m12, m13, m21, m22, m23, m31, m32, m33 float32) *ConvKernel {
	return &ConvKernel{
		Kernel: []float32{m11, m12, m13, m21, m22, m23, m31, m32, m33},
		Radius: 1, // 3x3 kernel has diameter=3, thus radius=1
	}
}

// Normalize the ConvKernel such that sum of all entries in the kernel matrix is 1. 
// If the current kernel entries sum to zero, no change is made.
// Modifies the current kernel.
func (k *ConvKernel) Normalize() {
	diameter := k.Radius*2 + 1
	area := diameter * diameter
	sum := float32(0)
	for i := 0; i < area; i++ {
		sum += k.Kernel[i]
	}

	// only attempt to normalize if the sum is significantly
	// different from both zero and one.
	fSum := float64(sum)
	if (math.Abs(fSum) >= TOLERANCE) && (math.Abs(fSum-1.0) >= TOLERANCE) { 
		for i := 0; i < area; i++ {
			k.Kernel[i] /= sum // normalize by dividing each entry
		}
	}
}

// a function for making sure a coord lies within plane bounds
// e.g. by clamping the co-ord or by wrapping it around the plane.
// input params are (index, max-index) ; output is the output-index.
type planeExtension func(int, int) int

// Clamp out-of-bounds pixels to nearest neighbour.
func clampPlaneExtension(index, limit int) int {
	if index >= limit {
		index = limit - 1
	} else if index < 0 {
		index = 0
	}
	return index
}

// edge wrapping: wrap out-of-bounds pixels around the image.
func wrapPlaneExtension(index, limit int) int { return index % limit }

// helper function for convolving a single intensity plane.
func convolvePlane(planePtr *[]float32, kernel *ConvKernel, width, height int, toPlaneCoords planeExtension) *[]float32 {

	plane := *planePtr
	radius := kernel.Radius
	diameter := radius*2 + 1
	res := make([]float32, width*height)

	// for each pixel of the intensity plane:
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*width + x

			// compute convolved value of the pixel:
			resV := float32(0)
			for yk := 0; yk < diameter; yk++ {
				yp := toPlaneCoords(y+yk-radius, height)
				for xk := 0; xk < diameter; xk++ {
					xp := toPlaneCoords(x+xk-radius, width)
					planeIndex := yp*width + xp
					kernelIndex := yk*diameter + xk
					resV += (plane[planeIndex] * kernel.Kernel[kernelIndex])
				}
			}
			res[index] = resV
		}
	}

	return &res
}

// Apply a convolution kernel to the image.
// Creates a new image (does not modify the original).
func (img *FloatImage) convolve(kernel *ConvKernel, px planeExtension) *FloatImage {

	// convolve each plane independently:
	res := new([3][]float32)
	for i := 0; i < 3; i++ {
		convolvePlane(&img.Ip[i], kernel, img.Width, img.Height, px)
	}

	return &FloatImage{
		Ip:     *res,
		Width:  img.Width,
		Height: img.Height,
	}
}

// Apply a convolution, in place, to the image.
// Modifies the current image.
func (img *FloatImage) convolveWith(kernel *ConvKernel, px planeExtension) {

	// convolve each plane independently:
	for i := 0; i < 3; i++ {
		img.Ip[i] = *convolvePlane(&img.Ip[i], kernel, img.Width, img.Height, px)
	}
}

// Apply a convolution kernel to the image, with Edge clamping.
// Creates a new image (does not modify the original).
func (img *FloatImage) ConvolveClamp(kernel *ConvKernel) *FloatImage {
	return img.convolve(kernel, clampPlaneExtension)
}

// Apply a convolution kernel to the image, with Edge wrapping.
// Creates a new image (does not modify the original).
func (img *FloatImage) ConvolveWrap(kernel *ConvKernel) *FloatImage {
	return img.convolve(kernel, wrapPlaneExtension)
}

// Apply a convolution, in place, to the image, with Edge clamping.
// Modifies the current image.
func (img *FloatImage) ConvolveClampWith(kernel *ConvKernel, px planeExtension) {
	img.convolveWith(kernel, clampPlaneExtension)
}

// Apply a convolution, in place, to the image, with Edge wrapping.
// Modifies the current image.
func (img *FloatImage) ConvolveWrapWith(kernel *ConvKernel, px planeExtension) {
	img.convolveWith(kernel, wrapPlaneExtension)
}

// a map function which operates on one pixel at a time
type PixelMap func(vals ...float32) float32

// Apply a PixelMap over each pixel over all images. 
// Modifies the current image.
// All images must have the same dimensions (this constraint is not checked).
func (img *FloatImage) Apply(mapFn PixelMap, images ...*FloatImage) {

	// obtain the number of args to the PixelMap fn
	numImages := len(images) + 1 // + 1 for the current image
	vals := make([]float32, numImages)

	for layer := 0; layer < 3; layer++ {
		for y := 0; y < img.Height; y++ {
			for x := 0; x < img.Width; x++ {
				index := y*img.Width + x

				// copy image pixels into vals
				vals[0] = img.Ip[layer][index]
				for i := 0; i < numImages; i++ {
					vals[i+1] = images[i].Ip[layer][index]
				}

				// apply the mapFunction
				img.Ip[layer][index] = mapFn(vals...)
			}
		}
	}
}

// Apply a PixelMap over each pixel over all images. 
// Does not modify the current image.
// All images must have the same dimensions (this constraint is not checked).
func Apply(mapFn PixelMap, images ...*FloatImage) *FloatImage {
	result := images[0].Clone()
	result.Apply(mapFn, images[1:]...)
	return result
}

