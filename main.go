package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xgraphics"
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
	outFlag := flag.String("out", "", "write glitch image to file as png")
	inFlag := flag.String("in", "", "read image from file")
	flag.Parse()

	if *debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	// use existing image
	if len(*inFlag) > 0 {
		file, err := os.Open(*inFlag)
		if err != nil {
			log.Panic(err)
		}
		img, _, err := image.Decode(file)
		if err != nil {
			log.Panic(err)
		}
		if img, ok := img.(*image.RGBA); ok {
			// img is now an *image.RGBA
			err = loop(img, *debugFlag, *passwordFlag)
			if err != nil {
				log.Panic(err)
			}
			return
		}
		log.Panic("input not an RGBA image")
	}

	if *piecesFlag <= 0 {
		log.Errorf("pieces must be > 0")
		return
	}
	screen, err := takeScreenshot()
	if err != nil {
		log.Panic(err)
	}
	if *censorFlag {
		screen, err = glitch.Censor(screen)
		if err != nil {
			log.Panic(err)
		}
	}
	screen, err = glitch.Distort(screen, &glitch.DistortConfig{
		Pixelate: *pixelateFlag,
		Pieces:   *piecesFlag,
		Seed:     *seedFlag,
	})
	if err != nil {
		log.Panic(err)
	}
	if len(*outFlag) > 0 {
		file, err := os.Create(*outFlag)
		if err != nil {
			log.Panic(err)
		}
		png.Encode(file, screen)
	}
	err = loop(screen, *debugFlag, *passwordFlag)
	if err != nil {
		log.Panic(err)
	}
}

func loop(screen *image.RGBA, permitEscape bool, customPassword string) error {
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
	// create window, set fullscreen, show image
	ximg := xgraphics.NewConvert(Xu, screen)
	win := ximg.XShow()
	err = ewmh.WmStateReq(Xu, win.Id, ewmh.StateToggle,
		"_NET_WM_STATE_FULLSCREEN")
	if err != nil {
		return err
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

func takeScreenshot() (*image.RGBA, error) {
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}
	return img, nil
}
