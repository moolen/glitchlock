# glitchlock
The glitchy X locker. Inspired by [xero/glitchlock](https://github.com/xero/glitchlock).

For additional "security" :trollface: glitchlock uses Tesseract's OCR engine to find characters on the screen and crosses them out. That's optional.

## Example

Example screenlock using `-censor`.

![glitchlock](https://i.imgur.com/kPwL42n.png)

## Installation

Grab a binary from the [releases page](https://github.com/moolen/glitchlock/releases) or `go get github.com/moolen/glitchlock` it. You need the tesseract development libraries for compiling this (`tesseract / archlinux` / `libtesseract-dev / ubuntu`) and for runtime, `tesseract-data-eng / archlinux` or `tesseract-ocr-eng / ubuntu`.

## Subpackages

* [PAM](https://github.com/moolen/glitchlock/blob/master/pam): check user/password combination using PAM.
* [glitch](https://github.com/moolen/glitchlock/blob/master/glitch): distort images.

## Known issues

* multi-head setup is WIP and might not work

## Usage

```
Usage of glitchlock:
  -censor
        censors text on the image
  -debug
        debug mode, hit ESC to exit
  -pieces int
        divices the screen into n pieces. Must be >0 (default 10)
  -pixelate int
        picelate width
  -seed int
        random seed (default 1312)
  -password string
        specify a custom unlock password. This ignores the user's password
  -version
        print version information and exit
```
