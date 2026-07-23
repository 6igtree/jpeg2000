# jpeg2000

[![Go Reference](https://pkg.go.dev/badge/github.com/d-fuji/jpeg2000.svg)](https://pkg.go.dev/github.com/d-fuji/jpeg2000)

English | [日本語](README.ja.md)

A pure Go decoder for JPEG 2000 (ISO/IEC 15444-1) images. No cgo, no
external libraries — the entire codec, from the MQ arithmetic coder to
the inverse wavelet transform, is implemented in Go.

```
go get github.com/d-fuji/jpeg2000
```

## Usage

The package plugs into the standard library's `image` package:

```go
import (
    "image"
    _ "github.com/d-fuji/jpeg2000"
)

f, _ := os.Open("photo.jp2")
img, _, err := image.Decode(f)
```

or call it directly:

```go
img, err := jpeg2000.Decode(f)
cfg, err := jpeg2000.DecodeConfig(f) // dimensions/color model only
```

Both the JP2 container format (`.jp2`) and raw codestreams
(`.j2k`/`.j2c`/`.jpc`) are supported.

## Features

- Reversible (5/3) and irreversible (9/7) wavelet transforms —
  lossless decoding is bit-exact
- Multiple tiles, tile-parts, quality layers, and resolution levels
- All five progression orders (LRCP, RLCP, RPCL, PCRL, CPRL)
- Custom precinct and code-block sizes, SOP/EPH markers
- RCT/ICT multiple-component (color) transforms
- Grayscale, RGB, and alpha-channel images, up to 16 bits per component
- Code-block styles: vertically causal contexts, predictable
  termination, termination on each pass, context reset, segmentation
  symbols

Not yet supported (see [ROADMAP.md](ROADMAP.md)): arithmetic coder
bypass mode, POC progression order changes, region of interest (RGN),
palettized images, component subsampling, HTJ2K (Part 15).

## Correctness

The test suite decodes images produced by OpenJPEG across the feature
matrix (transforms, progression orders, tilings, precincts, layers)
and compares against OpenJPEG's own decoded output: bit-exact for
lossless images, and within ±1 gray level (float rounding) for lossy
ones. The decoder is also fuzzed against malformed inputs.

To regenerate the test images (requires Python with Pillow):

```
python3 testdata/gen.py
```

## Status

This is a young library. The decoder is correct on everything in its
test matrix, but it is not yet optimized for speed and has not been
battle-tested against the full ISO conformance suite. Bug reports with
sample files are very welcome.

## License

Apache License 2.0. See [LICENSE](LICENSE).
