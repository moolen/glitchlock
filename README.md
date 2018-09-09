# glitchlock
The glitchy X locker. Inspired by [xero/glitchlock](https://github.com/xero/glitchlock).

For additional "security" :trollface: glitchlock uses Tesseract's OCR engine to find characters on the screen and crosses them out. That's optional.

## Example

Example screenlock using `-censor`.

![glitchlock](https://i.imgur.com/J3wi4Um.png)

For convenience, you can write the glitchy screenshot to a file `-out <filename>` and reuse it on the next run using `-in <filename>`. Further, you can specify a `-seed <int>` to reproduce glitch patterns.

## Subpackages

* [PAM](https://github.com/moolen/glitchlock/blob/master/pam): check user/password combination using PAM.
* [glitch](https://github.com/moolen/glitchlock/blob/master/glitch): distort images.

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
