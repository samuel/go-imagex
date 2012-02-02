package imagex

import (
	"image"
)

// Check entropy for every square skipping by granularity (really slow but accurate)
func SmartResizeAccurate(img image.Image, width, height, granularity int) image.Image {
	if granularity <= 0 {
		return nil
	}

	bnd := img.Bounds()
	iw, ih := bnd.Dx(), bnd.Dy()

	image_ratio := float64(iw) / float64(ih)
	size_ratio := float64(width) / float64(height)
	ratio := float64(iw*height) / float64(ih*width)
	abs_ratio := ratio
	if abs_ratio < 1.0 {
		abs_ratio = 1 / ratio
	}

	// If already just about matches ratio then just resize
	if abs_ratio < 1.001 {
		return Resize(img, img.Bounds(), width, height)
	}

	// Horizontal images
	if size_ratio < image_ratio {
		new_width := int(float64(ih) * size_ratio)
		maxx := 0
		maxe := float64(0)
		for x := 0; x < iw-new_width+1; x += granularity {
			if x >= iw-new_width+1 {
				x = iw - new_width
			}
			e := Entropy(img, image.Rect(x, 0, x+new_width, ih))
			if e > maxe {
				maxx = x
				maxe = e
			}
		}

		return Resize(img, image.Rect(maxx, 0, maxx+new_width, ih), width, height)
	}

	// Vertical images
	new_height := int(float64(iw) / size_ratio)
	maxy := 0
	maxe := float64(0)
	for y := 0; y < ih-new_height+1; y += granularity {
		if y >= ih-new_height+1 {
			y = ih - new_height
		}
		e := Entropy(img, image.Rect(0, y, iw, y+new_height))
		if e > maxe {
			maxy = y
			maxe = e
		}
	}

	return Resize(img, image.Rect(0, maxy, iw, maxy+new_height), width, height)
}

// Calculate entropy for stripes of the image and find largest sequence by sum
func SmartResizeStripes(img image.Image, width, height, granularity int) image.Image {
	if granularity <= 0 {
		return nil
	}

	bnd := img.Bounds()
	iw, ih := bnd.Dx(), bnd.Dy()

	image_ratio := float64(iw) / float64(ih)
	size_ratio := float64(width) / float64(height)
	ratio := float64(iw*height) / float64(ih*width)
	abs_ratio := ratio
	if abs_ratio < 1.0 {
		abs_ratio = 1 / ratio
	}

	// If already just about matches ratio then just resize
	if abs_ratio < 1.001 {
		return Resize(img, img.Bounds(), width, height)
	}

	// Horizontal images
	if size_ratio < image_ratio {
		new_width := int(float64(ih) * size_ratio)
		maxx := 0
		maxe := float64(0)
		entropies := make([]float64, iw/granularity+1)
		parts := new_width / granularity
		sume := float64(0)
		for x, i := (iw%granularity)/2, 0; x < iw; i += 1 {
			if x+granularity > iw {
				break
			}
			if i >= parts {
				sume -= entropies[i-parts]
			}
			e := Entropy(img, image.Rect(x, 0, x+granularity, ih))
			entropies[i] = e
			sume := e
			if i+1 >= parts && sume > maxe {
				maxe = sume
				maxx = x - (parts-1)*granularity
			}
			x += granularity
		}

		return Resize(img, image.Rect(maxx, 0, maxx+new_width, ih), width, height)
	}

	// Vertical images

	new_height := int(float64(iw) / size_ratio)
	maxy := 0
	maxe := float64(0)
	entropies := make([]float64, ih/granularity+1)
	parts := new_height / granularity
	sume := float64(0)
	for y, i := (ih%granularity)/2, 0; y < ih; i += 1 {
		if y+granularity > ih {
			break
		}
		if i >= parts {
			sume -= entropies[i-parts]
		}
		e := Entropy(img, image.Rect(0, y, iw, y+granularity))
		entropies[i] = e
		sume := e
		if i+1 >= parts && sume > maxe {
			maxe = sume
			maxy = y - (parts-1)*granularity
		}
		y += granularity
	}

	return Resize(img, image.Rect(0, maxy, iw, maxy+new_height), width, height)
}

// Check overflow broken up into pieces and remove on piece at a time (fast but inaccurate)
func SmartResizeTail(img image.Image, width, height, granularity int) image.Image {
	if granularity <= 0 {
		return nil
	}

	bnd := img.Bounds()
	iw, ih := bnd.Dx(), bnd.Dy()

	image_ratio := float64(iw) / float64(ih)
	size_ratio := float64(width) / float64(height)
	ratio := float64(iw*height) / float64(ih*width)
	abs_ratio := ratio
	if abs_ratio < 1.0 {
		abs_ratio = 1 / ratio
	}

	// If already just about matches ratio then just resize
	if abs_ratio < 1.001 {
		return Resize(img, img.Bounds(), width, height)
	}

	// Horizontal images
	if size_ratio < image_ratio {
		new_width := int(float64(ih) * size_ratio)
		left := 0
		right := iw
		el := float64(-1)
		er := float64(-1)
		for {
			dw := (right - left) - new_width
			if dw == 0 {
				break
			} else if dw > granularity {
				dw = granularity
			}

			if el < 0 {
				el = Entropy(img, image.Rect(left, 0, left+dw, ih))
				er = Entropy(img, image.Rect(right-dw, 0, right, ih))
			} else if el < er {
				er = Entropy(img, image.Rect(right-dw, 0, right, ih))
			} else {
				el = Entropy(img, image.Rect(left, 0, left+dw, ih))
			}
			if el < er {
				right -= dw
			} else {
				left += dw
			}
		}

		return Resize(img, image.Rect(left, 0, right, ih), width, height)
	}

	// Vertical images

	new_height := int(float64(iw) / size_ratio)
	top := 0
	bottom := ih
	et := float64(-1)
	eb := float64(-1)
	for {
		dh := (bottom - top) - new_height
		if dh == 0 {
			break
		} else if dh > granularity {
			dh = granularity
		}

		if et < 0 {
			et = Entropy(img, image.Rect(0, top, iw, top+dh))
			eb = Entropy(img, image.Rect(0, bottom-dh, iw, bottom))
		} else if et < eb {
			eb = Entropy(img, image.Rect(0, bottom-dh, iw, bottom))
		} else {
			et = Entropy(img, image.Rect(0, top, iw, top+dh))
		}
		if et < eb {
			bottom -= dh
		} else {
			top += dh
		}
	}

	return Resize(img, image.Rect(0, top, iw, bottom), width, height)
}
