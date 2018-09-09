# glitchlock
The glitchy X locker. Inspired by [xero/glitchlock](https://github.com/xero/glitchlock).

For additional security (:lol:) glitchlock uses Tesseract's OCR engine to find characters on the screen and crosses them out.

## Example

Example screenlock using `-censor`.
![glitchlock](https://i.imgur.com/J3wi4Um.png)

## Subpackages

[PAM](https://github.com/moolen/golock/blob/master/pam): check user/password combination using PAM.
[glitch](https://github.com/moolen/golock/blob/master/glitch): distort images.

## Usage

```
Usage of glitchlock:
  -censor
        censors text on the image
  -debug
        debug mode, hit ESC to exit
  -in string
        read image from file
  -out string
        write glitch image to file as png
  -pieces int
        divices the screen into n pieces. Must be >0 (default 10)
  -pixelate int
        picelate width
  -seed int
        random seed (default 1312)
```
