package main

import (
	"flag"
	"fmt"
	"image"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/kbinani/screenshot"
	"github.com/moolen/glitchlock/glitch"
	"github.com/moolen/glitchlock/pam"
	log "github.com/sirupsen/logrus"
)

func main() {
	censorFlag := flag.Bool("censor", false, "censors text on the image")
	seedFlag := flag.Int64("seed", 1312, "random seed")
	piecesFlag := flag.Int("pieces", 10, "divices the screen into n pieces. Must be >0")
	pixelateFlag := flag.Int("pixelate", 0, "picelate width")
	debugFlag := flag.Bool("debug", false, "debug mode, hit ESC to exit")
	passwordFlag := flag.String("password", "", "specify a custom unlock password. This ignores the user's password")
	flag.Parse()

	if *debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	if *piecesFlag <= 0 {
		log.Errorf("pieces must be > 0")
		return
	}
	screens, err := takeScreenshot()
	if err != nil {
		log.Panic(err)
	}
	screens, err = pipeline(screens, *censorFlag, *pixelateFlag, *piecesFlag, *seedFlag)
	if err != nil {
		log.Panic(err)
	}

	err = loop(screens, *debugFlag, *passwordFlag)
	if err != nil {
		log.Panic(err)
	}
}

type screen struct {
	image *image.RGBA
	rect  image.Rectangle
}

func (s screen) X() int {
	return s.rect.Min.X
}

func (s screen) Y() int {
	return s.rect.Min.Y
}

func (s screen) Width() int {
	return s.rect.Max.X - s.rect.Min.X
}

func (s screen) Height() int {
	return s.rect.Max.Y - s.rect.Min.Y
}

func pipeline(screens []*screen, censor bool, pixelate int, pieces int, seed int64) ([]*screen, error) {
	var err error
	for i, screen := range screens {
		if censor {
			screens[i].image, err = glitch.Censor(screen.image)
			if err != nil {
				return nil, err
			}
		}
		screens[i].image, err = glitch.Distort(screen.image, &glitch.DistortConfig{
			Pixelate: pixelate,
			Pieces:   pieces,
			Seed:     seed,
		})
		if err != nil {
			return nil, err
		}
	}
	return screens, nil
}

func loop(screens []*screen, permitEscape bool, customPassword string) error {
	// initialize xgb
	X, err := xgb.NewConn()
	if err != nil {
		return err
	}
	Xu, err := xgbutil.NewConnXgb(X)
	if err != nil {
		return err
	}
	keybind.Initialize(Xu)
	xscreen := xproto.Setup(X).DefaultScreen(X)
	// grab keyboard and pointer
	grabc := xproto.GrabKeyboard(X, false, xscreen.Root, xproto.TimeCurrentTime,
		xproto.GrabModeAsync, xproto.GrabModeAsync,
	)
	repk, err := grabc.Reply()
	if err != nil {
		return fmt.Errorf("error grabbing Keyboard")
	}
	if repk.Status != xproto.GrabStatusSuccess {
		return fmt.Errorf("could not grab keyboard")
	}
	grabp := xproto.GrabPointer(X, false, xscreen.Root, (xproto.EventMaskKeyPress|xproto.EventMaskKeyRelease)&0,
		xproto.GrabModeAsync, xproto.GrabModeAsync, xproto.WindowNone, xproto.CursorNone, xproto.TimeCurrentTime)
	repp, err := grabp.Reply()
	if err != nil {
		return fmt.Errorf("error grabbing pointer")
	}
	if repp.Status != xproto.GrabStatusSuccess {
		return fmt.Errorf("could not grab pointer")
	}

	for _, screen := range screens {
		ximg := xgraphics.NewConvert(Xu, screen.image)
		win, err := xwindow.Generate(ximg.X)
		if err != nil {
			return err
		}
		log.Debugf("creating window using screen rect %#v", screen.rect)
		win.Create(ximg.X.RootWin(), screen.X(), screen.Y(), screen.Width(), screen.Height(), 0)
		win.WMGracefulClose(func(w *xwindow.Window) {
			xevent.Detach(w.X, w.Id)
			keybind.Detach(w.X, w.Id)
			mousebind.Detach(w.X, w.Id)
			w.Destroy()
		})
		err = icccm.WmStateSet(ximg.X, win.Id, &icccm.WmState{
			State: icccm.StateNormal,
		})
		if err != nil { // not a fatal error
			return err
		}
		err = icccm.WmNormalHintsSet(ximg.X, win.Id, &icccm.NormalHints{
			Flags:     icccm.SizeHintPMinSize | icccm.SizeHintPMaxSize,
			MinWidth:  uint(screen.Width()),
			MinHeight: uint(screen.Height()),
			MaxWidth:  uint(screen.Width()),
			MaxHeight: uint(screen.Height()),
		})
		if err != nil {
			return err
		}

		// Paint our image before mapping.
		ximg.XSurfaceSet(win.Id)
		ximg.XDraw()
		ximg.XPaint(win.Id)
		win.Map()
	}

	// main loop
	lastInput := time.Now()
	var password string
	for {
		ev, err := X.WaitForEvent()
		if ev == nil && err == nil {
			return fmt.Errorf("Both event and error are nil. Exiting")
		}
		if err != nil {
			return err
		}
		if time.Now().Sub(lastInput) > time.Second*2 {
			log.Debugf("timeout reached. clearing password")
			password = ""
		}
		switch e := ev.(type) {
		case xproto.KeyPressEvent:
			key := keybind.LookupString(Xu, e.State, e.Detail)
			log.Debugf("keypress: %s %v ", key, e)
			lastInput = time.Now()
			if len(key) == 1 {
				password += key
			}
			log.Debugf("current password: %s", password)
			if keybind.KeyMatch(Xu, "Return", e.State, e.Detail) {
				log.Debugf("...checking password")
				if len(customPassword) > 0 {
					if password == customPassword {
						return nil
					}
					continue
				}
				if pam.AuthenticateCurrentUser(password) {
					return nil
				}
				log.Debugf("password does not match")
			}
			if permitEscape && keybind.KeyMatch(Xu, "Escape", e.State, e.Detail) {
				return nil
			}
		}
	}
}

func takeScreenshot() (out []*screen, err error) {
	var img *image.RGBA
	n := screenshot.NumActiveDisplays()
	log.Debugf("found %d screens", n)
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err = screenshot.CaptureRect(bounds)
		if err != nil {
			return
		}
		out = append(out, &screen{
			image: img,
			rect:  bounds,
		})
	}
	return
}
