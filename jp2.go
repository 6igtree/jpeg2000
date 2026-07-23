package jpeg2000

// JP2 container parsing (ISO/IEC 15444-1 Annex I): a sequence of boxes
// wrapping the raw codestream plus color/channel metadata.

const (
	boxSignature = 0x6A502020 // "jP  "
	boxFtyp      = 0x66747970 // "ftyp"
	boxJP2H      = 0x6A703268 // "jp2h"
	boxIHDR      = 0x69686472 // "ihdr"
	boxColr      = 0x636F6C72 // "colr"
	boxPclr      = 0x70636C72 // "pclr"
	boxCmap      = 0x636D6170 // "cmap"
	boxCdef      = 0x63646566 // "cdef"
	boxJP2C      = 0x6A703263 // "jp2c"
)

var jp2Signature = []byte{0, 0, 0, 0x0C, 0x6A, 0x50, 0x20, 0x20, 0x0D, 0x0A, 0x87, 0x0A}

// jp2Meta carries container-level information that affects rendering.
type jp2Meta struct {
	alphaChan int // component index of the opacity channel, -1 if none
}

func isJP2(data []byte) bool {
	return len(data) >= 12 && string(data[:12]) == string(jp2Signature)
}

func isRawCodestream(data []byte) bool {
	return len(data) >= 4 && be16(data) == mSOC && be16(data[2:]) == mSIZ
}

// extractCodestream returns the raw codestream from either a JP2 file
// or a bare codestream, along with container metadata.
func extractCodestream(data []byte) ([]byte, *jp2Meta, error) {
	meta := &jp2Meta{alphaChan: -1}
	if isRawCodestream(data) {
		return data, meta, nil
	}
	if !isJP2(data) {
		return nil, nil, errFormat("not a JPEG 2000 stream")
	}

	var codestream []byte
	var walk func(b []byte) error
	walk = func(b []byte) error {
		pos := 0
		for pos+8 <= len(b) {
			length := be32(b[pos:])
			btype := be32(b[pos+4:])
			hdr := 8
			switch length {
			case 0:
				length = len(b) - pos
			case 1:
				if pos+16 > len(b) {
					return errFormat("truncated box")
				}
				hi, lo := be32(b[pos+8:]), be32(b[pos+12:])
				if hi != 0 {
					return errFormat("box too large")
				}
				length = lo
				hdr = 16
			}
			if length < hdr || pos+length > len(b) {
				return errFormat("bad box length")
			}
			body := b[pos+hdr : pos+length]
			switch btype {
			case boxJP2H:
				if err := walk(body); err != nil {
					return err
				}
			case boxPclr, boxCmap:
				return UnsupportedError("palettized JP2 image")
			case boxCdef:
				if len(body) >= 2 {
					n := be16(body)
					for i := 0; i < n && 2+6*i+6 <= len(body); i++ {
						cn := be16(body[2+6*i:])
						typ := be16(body[2+6*i+2:])
						if typ == 1 { // opacity
							meta.alphaChan = cn
						}
					}
				}
			case boxJP2C:
				if codestream == nil {
					codestream = body
				}
			}
			pos += length
		}
		return nil
	}
	if err := walk(data); err != nil {
		return nil, nil, err
	}
	if codestream == nil {
		return nil, nil, errFormat("JP2 file has no codestream box")
	}
	return codestream, meta, nil
}
