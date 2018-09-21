package snap

import (
	"image"
	"image/color"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
)

type Screen struct {
	X, Y, Width, Height int
	Primary             bool
	Name                string
}

func (s Screen) Capture() (*image.RGBA, error) {
	c, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}
	defer c.Close()
	screen := xproto.Setup(c).DefaultScreen(c)
	wholeScreenBounds := image.Rect(0, 0, int(screen.WidthInPixels), int(screen.HeightInPixels))
	targetBounds := image.Rect(s.X, s.Y, s.X+s.Width, s.Y+s.Height)
	intersect := wholeScreenBounds.Intersect(targetBounds)
	rect := image.Rect(0, 0, s.Width, s.Height)
	img := image.NewRGBA(rect)
	if err != nil {
		return nil, err
	}

	var data []byte

	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root),
		int16(intersect.Min.X), int16(intersect.Min.Y),
		uint16(intersect.Dx()), uint16(intersect.Dy()), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data = xImg.Data

	offset := 0
	for iy := intersect.Min.Y; iy < intersect.Max.Y; iy++ {
		for ix := intersect.Min.X; ix < intersect.Max.X; ix++ {
			r := data[offset+2]
			g := data[offset+1]
			b := data[offset]
			img.SetRGBA(ix-(s.X), iy-(s.Y), color.RGBA{r, g, b, 255})
			offset += 4
		}
	}
	return img, nil
}

func GetScreens() (screens []Screen, err error) {
	var primaryOutput randr.Output
	X, err := xgb.NewConn()
	if err != nil {
		return
	}
	err = randr.Init(X)
	if err != nil {
		return
	}
	root := xproto.Setup(X).DefaultScreen(X).Root
	res, err := randr.GetScreenResources(X, root).Reply()
	if err != nil {
		return
	}
	primaryOutputReply, _ := randr.GetOutputPrimary(X, root).Reply()
	if primaryOutputReply != nil {
		primaryOutput = primaryOutputReply.Output
	}

	for _, output := range res.Outputs {
		oinfo, err := randr.GetOutputInfo(X, output, 0).Reply()
		if err != nil {
			return nil, err
		}
		if oinfo.Connection != randr.ConnectionConnected {
			continue
		}
		outputName := string(oinfo.Name)
		crtcinfo, err := randr.GetCrtcInfo(X, oinfo.Crtc, 0).Reply()
		if err != nil {
			return nil, err
		}
		screens = append(screens, Screen{
			Name:    outputName,
			X:       int(crtcinfo.X),
			Y:       int(crtcinfo.Y),
			Width:   int(crtcinfo.Width),
			Height:  int(crtcinfo.Height),
			Primary: output == primaryOutput,
		})
	}
	return
}
