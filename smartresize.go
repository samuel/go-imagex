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

	imageRatio := float64(iw) / float64(ih)
	sizeRatio := float64(width) / float64(height)
	ratio := float64(iw*height) / float64(ih*width)
	absRatio := ratio
	if absRatio < 1.0 {
		absRatio = 1 / ratio
	}

	// If already just about matches ratio then just resize
	if absRatio < 1.001 {
		return Resize(img, img.Bounds(), width, height)
	}

	// Horizontal images
	if sizeRatio < imageRatio {
		newWidth := int(float64(ih) * sizeRatio)
		maxx := 0
		maxe := float64(0)
		for x := 0; x < iw-newWidth+1; x += granularity {
			if x >= iw-newWidth+1 {
				x = iw - newWidth
			}
			e := Entropy(img, image.Rect(x, 0, x+newWidth, ih))
			if e > maxe {
				maxx = x
				maxe = e
			}
		}

		return Resize(img, image.Rect(maxx, 0, maxx+newWidth, ih), width, height)
	}

	// Vertical images
	newHeight := int(float64(iw) / sizeRatio)
	maxy := 0
	maxe := float64(0)
	for y := 0; y < ih-newHeight+1; y += granularity {
		if y >= ih-newHeight+1 {
			y = ih - newHeight
		}
		e := Entropy(img, image.Rect(0, y, iw, y+newHeight))
		if e > maxe {
			maxy = y
			maxe = e
		}
	}

	return Resize(img, image.Rect(0, maxy, iw, maxy+newHeight), width, height)
}

// Calculate entropy for stripes of the image and find largest sequence by sum
func SmartResizeStripes(img image.Image, width, height, granularity int) image.Image {
	if granularity <= 0 {
		return nil
	}

	bnd := img.Bounds()
	iw, ih := bnd.Dx(), bnd.Dy()

	imageRatio := float64(iw) / float64(ih)
	sizeRatio := float64(width) / float64(height)
	ratio := float64(iw*height) / float64(ih*width)
	absRatio := ratio
	if absRatio < 1.0 {
		absRatio = 1 / ratio
	}

	// If already just about matches ratio then just resize
	if absRatio < 1.001 {
		return Resize(img, img.Bounds(), width, height)
	}

	// Horizontal images
	if sizeRatio < imageRatio {
		newWidth := int(float64(ih) * sizeRatio)
		maxx := 0
		maxe := float64(0)
		entropies := make([]float64, iw/granularity+1)
		parts := newWidth / granularity
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

		return Resize(img, image.Rect(maxx, 0, maxx+newWidth, ih), width, height)
	}

	// Vertical images

	newHeight := int(float64(iw) / sizeRatio)
	maxy := 0
	maxe := float64(0)
	entropies := make([]float64, ih/granularity+1)
	parts := newHeight / granularity
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

	return Resize(img, image.Rect(0, maxy, iw, maxy+newHeight), width, height)
}

// Check overflow broken up into pieces and remove on piece at a time (fast but inaccurate)
func SmartResizeTail(img image.Image, width, height, granularity int) image.Image {
	if granularity <= 0 {
		return nil
	}

	bnd := img.Bounds()
	iw, ih := bnd.Dx(), bnd.Dy()

	imageRatio := float64(iw) / float64(ih)
	sizeRatio := float64(width) / float64(height)
	ratio := float64(iw*height) / float64(ih*width)
	absRatio := ratio
	if absRatio < 1.0 {
		absRatio = 1 / ratio
	}

	// If already just about matches ratio then just resize
	if absRatio < 1.001 {
		return Resize(img, img.Bounds(), width, height)
	}

	// Horizontal images
	if sizeRatio < imageRatio {
		newWidth := int(float64(ih) * sizeRatio)
		left := 0
		right := iw
		el := float64(-1)
		er := float64(-1)
		for {
			dw := (right - left) - newWidth
			if dw == 0 {
				break
			} else if dw > granularity {
				dw = granularity
			}

			switch {
			case el < 0:
				el = Entropy(img, image.Rect(left, 0, left+dw, ih))
				er = Entropy(img, image.Rect(right-dw, 0, right, ih))
			case el < er:
				er = Entropy(img, image.Rect(right-dw, 0, right, ih))
			default:
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

	newHeight := int(float64(iw) / sizeRatio)
	top := 0
	bottom := ih
	et := float64(-1)
	eb := float64(-1)
	for {
		dh := (bottom - top) - newHeight
		if dh == 0 {
			break
		} else if dh > granularity {
			dh = granularity
		}

		switch {
		case et < 0:
			et = Entropy(img, image.Rect(0, top, iw, top+dh))
			eb = Entropy(img, image.Rect(0, bottom-dh, iw, bottom))
		case et < eb:
			eb = Entropy(img, image.Rect(0, bottom-dh, iw, bottom))
		default:
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
