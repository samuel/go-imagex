package imagex

import (
	"image"
	"image/color"
)

func Histogram(img image.Image, rect image.Rectangle) ([]int, string) {
	var hist []int
	var pix []uint8
	var colormodel string
	var bounds image.Rectangle
	stride := 0
	n := 0
	switch img := img.(type) {
	case *image.RGBA:
		colormodel = "rgba"
		img = img.SubImage(rect).(*image.RGBA)
		pix = img.Pix
		n = 4
		stride = img.Stride
		bounds = img.Bounds()
	case *image.NRGBA:
		colormodel = "rgba"
		img = img.SubImage(rect).(*image.NRGBA)
		pix = img.Pix
		n = 4
		stride = img.Stride
		bounds = img.Bounds()
	// case *image.RGBA64: -- 16-bit values
	case *image.Gray:
		colormodel = "gray"
		img = img.SubImage(rect).(*image.Gray)
		pix = img.Pix
		n = 1
		stride = img.Stride
		bounds = img.Bounds()
	// case *image.Gray16: -- 16-bit values
	case *image.YCbCr:
		colormodel = "ycbcr"
		hist = HistogramYCbCr(img, rect, false)
		if hist != nil {
			return hist, colormodel
		}
	default:
		colormodel = "other"
	}
	if n > 0 {
		hist = make([]int, n*256)
		for y, o := bounds.Min.Y, 0; y < bounds.Max.Y; y, o = y+1, o+stride {
			for x := 0; x < n*bounds.Dx(); x += n {
				for i := 0; i < n; i++ {
					hist[int(pix[o+x+i])+i*256] += 1
				}
			}
		}
	} else {
		colormodel = "rgb"
		hist = make([]int, 3*256)
		for y := rect.Min.Y; y < rect.Max.Y; y += 1 {
			for x := rect.Min.X; x < rect.Max.X; x += 1 {
				color := img.At(x, y)
				r, g, b, _ := color.RGBA()
				hist[r>>8] += 1
				hist[(g>>8)+256] += 1
				hist[(b>>8)+512] += 1
			}
		}
	}
	return hist, colormodel
}

func HistogramYCbCr(img *image.YCbCr, rect image.Rectangle, rgb bool) []int {
	var verticalRes int
	var horizontalRes int
	switch img.SubsampleRatio {
	case image.YCbCrSubsampleRatio420:
		verticalRes = 2
		horizontalRes = 2
	case image.YCbCrSubsampleRatio422:
		verticalRes = 1
		horizontalRes = 2
	case image.YCbCrSubsampleRatio444:
		verticalRes = 1
		horizontalRes = 1
	default:
		return nil
	}

	hist := make([]int, 256*3)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		Y := img.Y[y*img.YStride:]
		Cb := img.Cb[y/verticalRes*img.CStride:]
		Cr := img.Cr[y/verticalRes*img.CStride:]
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if rgb {
				r, g, b := color.YCbCrToRGB(Y[x], Cb[x/horizontalRes], Cr[x/horizontalRes])
				hist[r] += 1
				hist[int(g)+256] += 1
				hist[int(b)+512] += 1
			} else {
				hist[Y[x]] += 1
				hist[int(Cb[x/horizontalRes])+256] += 1
				hist[int(Cr[x/horizontalRes])+512] += 1
			}
		}
	}

	return hist
}
