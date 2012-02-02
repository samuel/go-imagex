package imagex

import (
	"image"
	"math"
)

func Entropy(img image.Image, rect image.Rectangle) float64 {
	hist, cm := Histogram(img, rect)
	switch cm {
	case "rgba":
		hist = hist[:768] // Ignore alpha channel
	case "ycbcr":
		hist = hist[:256] // Only use Y channel
	}

	hist_len := 0
	for i := 0; i < len(hist); i++ {
		hist_len += hist[i]
	}

	e := float64(0)
	for _, v := range hist {
		if v != 0 {
			p := float64(v) / float64(hist_len)
			e += p * math.Log2(p)
		}
	}

	return -e
}
