// Implements image sharpening techniques.
package imgproc

import "math"

// Sharpen the image with the given amount, radius and threshold, using
// the Unsharp mask technique.
// mutates the current image.
func (img *FloatImage) Unsharp(radius int, amount, threshold float64) {

	// TODO - possibly convert to HSV, apply transform on value only, convert back

	// apply gaussian blur
	gaussKernel := GaussianFilterKernel(radius, amount)
	blurImg := img.ConvolveClamp(gaussKernel)

	// if diff(orig, blur) > threshold, apply subtraction
	unsharpFn := func(vals ...float32) float32 {
		orig, blur := vals[0], vals[1]
		if diff := orig - blur; math.Abs(float64(diff)) > threshold {
			return orig + diff
		}
		return orig
	}

	img.Apply(unsharpFn, blurImg)
}

// Sharpen the image as per FloatImage.Unsharp, except return a new image 
// rather than modifying the original image.
func Unsharp(img *FloatImage, radius int, amount, threshold float64) *FloatImage {
	result := img.Clone() // init new image
	result.Unsharp(radius, amount, threshold)
	return result
}

// Sharpen the image using the Laplacian.
// Modifies the current image.
func (img *FloatImage) SharpenLaplace() {
	// add laplacian to the image
	laplacian := img.ConvolveClamp(LaplaceSpherical())
	img.Apply(func(v ...float32) float32 { return v[0] + v[1] }, laplacian)
	// TODO histogram equalisation, to compensate for the increased brightness.
}

// Sharpen the image using the Laplacian.
// Returns a new image (rather than modifying the current image).
func SharpenLaplace(img *FloatImage) *FloatImage {
	result := img.Clone() // init new image
	result.SharpenLaplace()
	return result
}
