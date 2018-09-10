package glitch

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math/rand"

	"github.com/disintegration/gift"
	"github.com/moolen/gosseract"
)

// DistortConfig configures the glitch behavior
type DistortConfig struct {
	Pixelate int
	Pieces   int
	Seed     int64
}

// Distort distorts the input image by manipulating color channels and cropping/rotating/translating slices of the input image
func Distort(in *image.RGBA, dc *DistortConfig) (*image.RGBA, error) {

	rand.Seed(dc.Seed)
	// 1st pass: shift colors and distort
	dst := image.NewRGBA(in.Bounds())
	gft := gift.New(
		gift.Pixelate(dc.Pixelate),
		gift.ColorFunc(
			func(r0, g0, b0, a0 float32) (r, g, b, a float32) {
				r = 1 - r0
				g = g0 + 0.1
				b = 0
				a = a0
				return
			},
		),
		gift.Convolution(
			[]float32{
				-1, -1, -1,
				-1, 1, 1,
				2, 1, 1,
			},
			false, true, false, 0.0,
		),
		gift.Sepia(60),
	)
	gft.Draw(dst, in)

	// 2nd pass: slice image into pieces and iterate over them
	bounds := dst.Bounds()
	pcs := dc.Pieces
	sliceHeight := bounds.Max.Y / pcs
	var currentMin, currentMax int
	for i := 1; i <= pcs; i++ {
		opt := rand.Intn(i * 3)
		heightV := rand.Intn(sliceHeight * 2)
		Xoffset := rand.Intn(sliceHeight)
		currentMin = (i - 1) * sliceHeight
		currentMax = i*sliceHeight + heightV
		g := gift.New(gift.Crop(image.Rect(bounds.Min.X, currentMin, bounds.Max.X, currentMax)))

		// shift colors
		if (i-opt)%4 == 0 {
			g.Add(gift.ColorFunc(
				func(r0, g0, b0, a0 float32) (r, g, b, a float32) {
					r = 1 - r0
					g = g0 - 0.3
					b = r
					a = a0
					return
				},
			))
		}
		if (i-opt)%3 == 0 {
			g.Add(gift.ColorFunc(
				func(r0, g0, b0, a0 float32) (r, g, b, a float32) {
					r = g0
					g = g0 + 0.4
					b = r
					a = a0
					return
				},
			))
		}
		// flip slice
		if (i+opt)%6 == 0 {
			g.Add(gift.FlipHorizontal())
		}
		if (i+opt)%5 == 0 {
			g.Add(gift.FlipVertical())
		}
		g.Add(gift.Sepia(30))
		g.DrawAt(dst, dst, image.Point{bounds.Min.X + Xoffset, currentMin}, gift.OverOperator)
	}

	return dst, nil
}

// Censor analyzes the input image for text and censors it
func Censor(in *image.RGBA) (*image.RGBA, error) {
	client := gosseract.NewClient()
	defer client.Close()
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, in)
	client.SetImageFromBytes(pngBuf.Bytes())
	coords, _ := client.GetResults()
	for _, coord := range coords {
		in = drawRect(in, coord.X1, coord.Y1, coord.X2, coord.Y2)
	}
	return in, nil
}

func drawRect(img *image.RGBA, x1, y1, x2, y2 int) *image.RGBA {
	for y := y1; y < y2; y++ {
		for x := x1; x <= x2; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	return img
}
