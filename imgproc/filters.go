package imgproc

import "math"

// build up a NxN matrix, populated by zeros
func emptyKernel(radius int) (area, diameter int, kernel []float32) {
	diameter = 2*radius + 1
	area = diameter * diameter
	kernel = make([]float32, area)
	return
}

// a mean filter: averages each pixel across it's neighbourhood.
// The neighbourhood is a NxN matrix, where N is 2*radius+1.
func MeanFilterKernel(radius int) *ConvKernel {

	// populate the matrix with 1/(N*N) values
	area, _, kernel := emptyKernel(radius)
	mean := float32(1) / float32(area)
	for i := 0; i < area; i++ {
		kernel[i] = mean
	}

	return &ConvKernel{
		Kernel: kernel,
		Radius: radius,
	}
}

// a gaussian filter: averages each pixel using the neighbourhood,
//  weighted by the (sampled) Gaussian function.
// The neighbourhood is a NxN matrix, where N is 2*radius+1.
// variance is the variance of the Gaussian (i.e. sigma squared).
func GaussianFilterKernel(radius int, variance float64) *ConvKernel {

	// Gaussian function (below comments are in \LaTeX math notation):
	// G(x,y) = \frac{1}{2\pi\sigma^2} e^{-\frac{x^2+y^2}{2\sigma^2}}
	// Let \alpha = \frac{1}{2\sigma^2}
	// Let \beta = \frac{1}{2\pi\sigma^2} = \frac{\alpha}{\pi}
	// Then G(x,y) = \beta e^{-\alpha (x^2+y^2)}
	alpha := 0.5 / variance
	beta := alpha / math.Pi

	// populate the matrix with the Gaussian function
	// exploiting the symmetry in the four directions
	_, diameter, kernel := emptyKernel(radius)
	for x := 0; x <= radius; x++ {
		for y := 0; y <= radius; y++ {
			// compute Gaussian weighting:
			exp := -alpha * float64(x*x+y*y)
			gauss := float32(beta * math.Exp(exp))

			// The 4 pixels with the same weighting are:
			//  (x,y), (-x,y), (x,-y), (-x, -y)
			// convert into kernel matrix indices by shifting by radius
			x1, x2 := radius+x, radius-x
			y1, y2 := radius+y, radius-y
			kernel[y1*diameter+x1] = gauss
			kernel[y1*diameter+x2] = gauss
			kernel[y2*diameter+x1] = gauss
			kernel[y2*diameter+x2] = gauss
		}
	}

	// Normalize the kernel before returning
	res := &ConvKernel{
		Kernel: kernel,
		Radius: radius,
	}
	res.Normalize()
	return res

}

// laplacian operator: without diagonals
// 0  1  0 
// 1 -4  1 
// 0  1  0
func LaplaceWithoutDiagonal() *ConvKernel {
	c, m, o := float32(0), float32(1), float32(-4) // corner, middle, origin
	return NewConvKernel3(c, m, c, m, o, m, c, m, c)
}

// laplacian operator: with diagonals
// 0.5  1.0  0.5 
// 1.0 -6.0  1.0 
// 0.5  1.0  0.5
func LaplaceWithDiagonal() *ConvKernel {
	c, m, o := float32(0.5), float32(1), float32(-6) // corner, middle, origin
	return NewConvKernel3(c, m, c, m, o, m, c, m, c)
}

// laplacian operator: with diagonals equally weighted with adjacent neighbours
// 1  1  1 
// 1 -8  1 
// 1  1  1
func LaplaceSpherical() *ConvKernel {
	n, o := float32(1), float32(-8) // neighbour, origin
	return NewConvKernel3(n, n, n, n, o, n, n, n, n)
}
