# Roadmap

English | [日本語](ROADMAP.ja.md)

The goal of this project is a complete, fast, dependency-free JPEG 2000
codec for Go. Development proceeds decoder-first: each milestone keeps
the library releasable and the test matrix green.

## v0.1 — Baseline decoder (done)

- [x] JP2 container and raw codestream parsing
- [x] Reversible 5/3 and irreversible 9/7 inverse wavelet transforms
- [x] MQ arithmetic decoder and EBCOT Tier-1 coding passes
- [x] Packet headers, tag trees, all five progression orders
- [x] Multiple tiles, tile-parts, quality layers, precincts
- [x] RCT/ICT color transforms, alpha channels, 8/16-bit output
- [x] `image.Decode` registration, `DecodeConfig`
- [x] Reference tests against OpenJPEG output; fuzz harness

## v0.2 — Part 1 completeness

Fill in the remaining conformance features so any spec-compliant
Part 1 file decodes:

- [ ] Arithmetic coder bypass ("lazy") mode and raw codeword segments
- [ ] POC marker / progression order changes
- [ ] Region of interest (RGN marker, ROI max-shift)
- [ ] Packed packet headers (PPM/PPT markers)
- [ ] Palettized images (pclr/cmap boxes) and full cdef handling
- [ ] Component subsampling (chroma-subsampled codestreams)
- [ ] ICC profile extraction (expose via decode options)
- [ ] Error resilience: tolerate truncated codestreams by returning
      the partial image instead of an error
- [ ] Validate against the ISO/IEC 15444-4 conformance test suite

## v0.3 — Performance

Target: within ~2x of OpenJPEG single-threaded, and faster
multi-threaded:

- [ ] Profile and optimize the Tier-1 inner loops (flag-based
      neighborhood state instead of recomputed contexts)
- [ ] In-place lifting DWT without per-column copies
- [ ] Concurrent decoding: goroutine per tile and per code-block
- [ ] Reduce allocations (buffer reuse across code-blocks)
- [ ] Benchmarks in CI to catch regressions

## v0.4 — Progressive & partial decoding

Exploit JPEG 2000's scalability — the main reason to choose the format:

- [ ] Decode options API: `DecodeWithOptions(r, &Options{...})`
- [ ] Reduced-resolution decode (stop after N resolution levels)
- [ ] Quality-layer truncation (decode only the first N layers)
- [ ] Region decode (only the tiles/precincts covering a rectangle)
- [ ] Streaming input: decode from an `io.Reader` without buffering
      the whole file

## v0.5 — Encoder

- [ ] Lossless encoding (5/3, RCT, single layer)
- [ ] Lossy encoding (9/7, ICT, rate control with PCRD-opt)
- [ ] Tiling, progression order and precinct configuration
- [ ] `image.Image` → JP2 and raw codestream writers
- [ ] Round-trip tests and rate-distortion comparison vs OpenJPEG

## v0.6 — HTJ2K (Part 15)

High-throughput JPEG 2000 (JPH) replaces the MQ coder with a much
faster block coder and is where the ecosystem is heading:

- [ ] HT block decoder (cleanup/sigprop/magref segments)
- [ ] Mixed HT/legacy code-block handling
- [ ] HT encoding

## v1.0 — Stability

- [ ] API freeze after real-world feedback
- [ ] Full conformance documentation (which class/profile is met)
- [ ] Continuous fuzzing (OSS-Fuzz integration)
- [ ] Memory-bounded decoding guarantees for untrusted input

## Non-goals (for now)

- JPIP (Part 9 interactive protocol)
- MJ2 motion JPEG 2000 (Part 3)
- Part 2 extensions (arbitrary wavelet kernels, multiple component
  transform extensions)

Contributions toward any milestone are welcome — file an issue with a
sample image if you hit an unsupported feature.
