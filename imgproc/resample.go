package imgproc

// Lerp linearly interpolates between two values, at x0 and x2.
// The value at x0 is f0 (short for "f(x0)") and the value at x2 is f2.
// Lerp requires that x0 <= x1 <= x2 and x0 < x2.
// Lerp returns the interpolated value for x1, that is, f1.
func lerp(x0, x1, x2, f0, f2 float32) float32 {

	// weight contribution by the other's dist
	x0dist, x2dist := x1-x0, x2-x1
	return (f0*x2dist + f2*x0dist) / (x0dist + x2dist)
}

// Bilerp bilinearly interpolates between 4 points, at (x0,y0), (x0,y2), (x2,y0), and (x2,y2).
// The respective values of the points are f00 (short for "f(x0,y0)"), f02, f20, and f22.
// Bilerp requires x0 <= x1 <= x2 with x0 < x2 and y0 <= y1 <= y2 with y0 < y2.
// Bilerp computes the value of the point (x1,y1), i.e. returns f11.
func bilerp(x0, x1, x2, y0, y1, y2, f00, f02, f20, f22 float32) float32 {
	// lerp in x-dir
	f10 := lerp(x0, x1, x2, f00, f20)
	f12 := lerp(x0, x1, x2, f20, f22)
	// then, lerp in y-dir
	return lerp(y0, y1, y2, f10, f12)
}

// CubicInterpolation interpolates a point (x2) in a single dimension, using cubic interpolation.
// Requires x0 < x1 <= x2 <= x3 < x4, with x4 - x3 = x3 - x1 = x1 - x0 > 0.
// (I.e. x0,x1,x3,x4 must be 4 distinct, equally spaced points. x2 must lie in between x1 and x3).
// The values of the points are f0 (short for "f(x0)"), f1, f3, and f4. 
// CubicInterpolation computes the value of x2, i.e. returns f2.
func cubicInterpolation(x0, x1, x2, x3, x4, f0, f1, f3, f4 float32) float32 {
	// compute where x2 lies in between x1 and x3:
	t := (x2 - x1) / (x3 - x1)

	// use precomputed formula.
	// Wikipedia [http://en.wikipedia.org/wiki/Bicubic_interpolation#Bicubic_convolution_algorithm], 
	// Reference: 
	//  R. Keys, (1981). 
	//  "Cubic convolution interpolation for digital image processing". 
	//  IEEE Transactions on Signal Processing, Acoustics, Speech, and Signal Processing
	res := 3*(f1-f3) - f0 + f4 // t^3 coef (short for "coefficient")
	res *= t
	res += 2*f0 - 5*f1 + 4*f3 - f4 // t^2 coef
	res *= t
	res += f3 - f0 // t coef
	res *= float32(0.5) * t
	res += f1 // const coef
	return res
}

// BicubicInterpolation interpolates a point (x2,y2) using a 4x4 grid (i.e. of 16 points)
// Requires x0 < x1 <= x2 <= x3 < x4, with x4 - x3 = x3 - x1 = x1 - x0 > 0.
// (I.e. x0,x1,x3,x4 must be 4 distinct, equally spaced points. x2 must lie in between x1 and x3).
// Similarly, for y: y0 < y1 <= y2 <= y3 < y4 with y4 - y4 = y3 - y1 = y1 - y0 > 0.
// The values of the points are e.g. f00, which is short for "f(x0,y0)".
// BicubicInterpolation computes the value of the (x2,y2) point, i.e. returns f22.
func bicubicInterpolation(
	x0, x1, x2, x3, x4,
	y0, y1, y2, y3, y4,
	f00, f01, f03, f04,
	f10, f11, f13, f14,
	f30, f31, f33, f34,
	f40, f41, f43, f44 float32) float32 {

	// interpolate in x-dir:
	f20 := cubicInterpolation(x0, x1, x2, x3, x4, f00, f10, f30, f40)
	f21 := cubicInterpolation(x0, x1, x2, x3, x4, f01, f11, f31, f41)
	f23 := cubicInterpolation(x0, x1, x2, x3, x4, f03, f13, f33, f43)
	f24 := cubicInterpolation(x0, x1, x2, x3, x4, f04, f14, f34, f44)

	// then, interpolate in the y-dir:
	return cubicInterpolation(y0, y1, y2, y3, y4, f20, f21, f23, f24)
}
