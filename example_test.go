package jpeg2000_test

import (
	"fmt"
	"image"
	"os"

	_ "github.com/d-fuji/jpeg2000"
)

// Importing the package registers the format, so the standard
// image.Decode recognizes JPEG 2000 files.
func Example() {
	f, err := os.Open("testdata/rgb_lossless.jp2")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	b := img.Bounds()
	fmt.Printf("%s %dx%d\n", format, b.Dx(), b.Dy())
	// Output: jp2 97x61
}
