# glitch



## Example

See [glitchlock](https://github.com/moolen/glitchlock).


```go
package main

import (
        "image/png"
        "os"

        "github.com/kbinani/screenshot"
        "github.com/moolen/glitchlock/glitch"
)

func main() {
        bounds := screenshot.GetDisplayBounds(0)
        img, _ := screenshot.CaptureRect(bounds)

        // first censor, then distort
        censored, err := glitch.Censor(img)
        if err != nil {
                panic(err)
        }
        glitch, err := glitch.Distort(censored, &glitch.DistortConfig{
                Pixelate: 3,
                Pieces:   10,
                Seed:     1312,
        })
        if err != nil {
                panic(err)
        }
        file, err := os.Create("glitch.png")
        if err != nil {
                panic(err)
        }
        defer file.Close()
        png.Encode(file, glitch)
}

```
