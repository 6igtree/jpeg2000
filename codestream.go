package jpeg2000

import "encoding/binary"

// Marker codes (ISO/IEC 15444-1 Table A.2).
const (
	mSOC = 0xFF4F
	mSIZ = 0xFF51
	mCOD = 0xFF52
	mCOC = 0xFF53
	mTLM = 0xFF55
	mPLM = 0xFF57
	mPLT = 0xFF58
	mQCD = 0xFF5C
	mQCC = 0xFF5D
	mRGN = 0xFF5E
	mPOC = 0xFF5F
	mPPM = 0xFF60
	mPPT = 0xFF61
	mCRG = 0xFF63
	mCOM = 0xFF64
	mSOT = 0xFF90
	mSOP = 0xFF91
	mEPH = 0xFF92
	mSOD = 0xFF93
	mEOC = 0xFFD9
)

// Progression orders (Table A.16).
const (
	progLRCP = 0
	progRLCP = 1
	progRPCL = 2
	progPCRL = 3
	progCPRL = 4
)

// Code-block style bits (Table A.19).
const (
	cbBypass     = 0x01
	cbResetCtx   = 0x02
	cbTermAll    = 0x04
	cbVertCausal = 0x08
	cbPredTerm   = 0x10
	cbSegSym     = 0x20
)

type sizComponent struct {
	depth  int
	signed bool
	dx, dy int
}

type sizInfo struct {
	xsiz, ysiz     int
	xosiz, yosiz   int
	xtsiz, ytsiz   int
	xtosiz, ytosiz int
	comps          []sizComponent
}

// codingParams holds the COD/COC parameters that apply per component.
type codingParams struct {
	numDecomps int
	cbW, cbH   uint // code-block size exponents
	cbStyle    uint8
	transform  uint8 // 0: irreversible 9/7, 1: reversible 5/3
	precW      []uint
	precH      []uint // precinct size exponents, one per resolution
}

// codHeader is a full COD marker: tile-wide settings plus the default
// per-component coding parameters.
type codHeader struct {
	progression uint8
	numLayers   int
	mct         bool
	sop, eph    bool
	cp          codingParams
}

type quantParams struct {
	style     int // 0: none, 1: scalar derived, 2: scalar expounded
	guardBits int
	exps      []int
	mants     []int
}

type tileData struct {
	index int
	data  []byte
	cod   *codHeader
	cocs  []*codingParams
	qcd   *quantParams
	qccs  []*quantParams
}

type codestream struct {
	siz                  sizInfo
	numXTiles, numYTiles int
	cod                  codHeader
	cocs                 []*codingParams
	qcd                  quantParams
	qccs                 []*quantParams
	tiles                []*tileData
}

// codingFor resolves the coding parameters for a component of a tile,
// applying the marker precedence: tile COC > tile COD > main COC > main COD.
func (cs *codestream) codingFor(t *tileData, comp int) *codingParams {
	if t.cocs[comp] != nil {
		return t.cocs[comp]
	}
	if t.cod != nil {
		return &t.cod.cp
	}
	if cs.cocs[comp] != nil {
		return cs.cocs[comp]
	}
	return &cs.cod.cp
}

func (cs *codestream) quantFor(t *tileData, comp int) *quantParams {
	if t.qccs[comp] != nil {
		return t.qccs[comp]
	}
	if t.qcd != nil {
		return t.qcd
	}
	if cs.qccs[comp] != nil {
		return cs.qccs[comp]
	}
	return &cs.qcd
}

func (cs *codestream) tileHeader(t *tileData) *codHeader {
	if t.cod != nil {
		return t.cod
	}
	return &cs.cod
}

func be16(b []byte) int { return int(binary.BigEndian.Uint16(b)) }
func be32(b []byte) int { return int(binary.BigEndian.Uint32(b)) }

func ceilDiv(a, b int) int { return (a + b - 1) / b }

// parseCodestream parses a raw JPEG 2000 codestream (SOC..EOC) into
// its headers and per-tile compressed data.
func parseCodestream(data []byte) (*codestream, error) {
	if len(data) < 4 || be16(data) != mSOC {
		return nil, errFormat("missing SOC marker")
	}
	cs := &codestream{}
	pos := 2
	seenSIZ, seenCOD, seenQCD := false, false, false

	readSeg := func() (marker int, body []byte, err error) {
		if pos+2 > len(data) {
			return 0, nil, errFormat("truncated codestream")
		}
		marker = be16(data[pos:])
		pos += 2
		if marker == mEOC || marker == mSOD || marker == mSOT {
			return marker, nil, nil
		}
		if pos+2 > len(data) {
			return 0, nil, errFormat("truncated marker segment")
		}
		l := be16(data[pos:])
		if l < 2 || pos+l > len(data) {
			return 0, nil, errFormat("bad marker segment length")
		}
		body = data[pos+2 : pos+l]
		pos += l
		return marker, body, nil
	}

	// Main header.
mainHeader:
	for {
		marker, body, err := readSeg()
		if err != nil {
			return nil, err
		}
		switch marker {
		case mSIZ:
			if err := parseSIZ(body, &cs.siz); err != nil {
				return nil, err
			}
			n := len(cs.siz.comps)
			cs.cocs = make([]*codingParams, n)
			cs.qccs = make([]*quantParams, n)
			cs.numXTiles = ceilDiv(cs.siz.xsiz-cs.siz.xtosiz, cs.siz.xtsiz)
			cs.numYTiles = ceilDiv(cs.siz.ysiz-cs.siz.ytosiz, cs.siz.ytsiz)
			cs.tiles = make([]*tileData, cs.numXTiles*cs.numYTiles)
			seenSIZ = true
		case mCOD:
			if !seenSIZ {
				return nil, errFormat("COD before SIZ")
			}
			if err := parseCOD(body, &cs.cod); err != nil {
				return nil, err
			}
			seenCOD = true
		case mCOC:
			if !seenSIZ {
				return nil, errFormat("COC before SIZ")
			}
			comp, cp, err := parseCOC(body, len(cs.siz.comps), &cs.cod.cp)
			if err != nil {
				return nil, err
			}
			cs.cocs[comp] = cp
		case mQCD:
			if err := parseQCD(body, &cs.qcd); err != nil {
				return nil, err
			}
			seenQCD = true
		case mQCC:
			comp, qp, err := parseQCC(body, len(cs.siz.comps))
			if err != nil {
				return nil, err
			}
			cs.qccs[comp] = qp
		case mPOC:
			return nil, UnsupportedError("POC progression order change")
		case mRGN:
			return nil, UnsupportedError("RGN region of interest")
		case mPPM:
			return nil, UnsupportedError("PPM packed packet headers")
		case mTLM, mPLM, mCRG, mCOM:
			// Informational; skip.
		case mSOT:
			pos -= 2
			break mainHeader
		case mEOC:
			return nil, errFormat("EOC before any tile data")
		default:
			// Unknown but well-formed marker segment; skip.
		}
	}
	if !seenSIZ || !seenCOD || !seenQCD {
		return nil, errFormat("main header missing SIZ, COD or QCD")
	}

	// Tile parts.
	for {
		if pos+2 > len(data) {
			return nil, errFormat("truncated codestream (no EOC)")
		}
		marker := be16(data[pos:])
		if marker == mEOC {
			break
		}
		if marker != mSOT {
			return nil, errFormat("expected SOT marker")
		}
		if pos+12 > len(data) {
			return nil, errFormat("truncated SOT")
		}
		isot := be16(data[pos+4:])
		psot := be32(data[pos+6:])
		sotStart := pos
		pos += 12
		if isot >= len(cs.tiles) {
			return nil, errFormat("tile index out of range")
		}
		t := cs.tiles[isot]
		if t == nil {
			n := len(cs.siz.comps)
			t = &tileData{
				index: isot,
				cocs:  make([]*codingParams, n),
				qccs:  make([]*quantParams, n),
			}
			cs.tiles[isot] = t
		}

		// Tile-part header markers until SOD.
	tilePartHeader:
		for {
			marker, body, err := readSeg()
			if err != nil {
				return nil, err
			}
			switch marker {
			case mSOD:
				break tilePartHeader
			case mCOD:
				h := &codHeader{}
				if err := parseCOD(body, h); err != nil {
					return nil, err
				}
				t.cod = h
			case mCOC:
				base := &cs.cod.cp
				if t.cod != nil {
					base = &t.cod.cp
				}
				comp, cp, err := parseCOC(body, len(cs.siz.comps), base)
				if err != nil {
					return nil, err
				}
				t.cocs[comp] = cp
			case mQCD:
				qp := &quantParams{}
				if err := parseQCD(body, qp); err != nil {
					return nil, err
				}
				t.qcd = qp
			case mQCC:
				comp, qp, err := parseQCC(body, len(cs.siz.comps))
				if err != nil {
					return nil, err
				}
				t.qccs[comp] = qp
			case mPOC:
				return nil, UnsupportedError("POC progression order change")
			case mRGN:
				return nil, UnsupportedError("RGN region of interest")
			case mPPT:
				return nil, UnsupportedError("PPT packed packet headers")
			case mPLT, mCOM:
				// Skip.
			case mEOC:
				return nil, errFormat("EOC inside tile-part header")
			default:
			}
		}

		var end int
		if psot == 0 {
			// Last tile-part: extends to EOC.
			if len(data) < 2 {
				return nil, errFormat("truncated codestream")
			}
			end = len(data) - 2
		} else {
			end = sotStart + psot
		}
		if end < pos || end > len(data) {
			return nil, errFormat("bad Psot tile-part length")
		}
		t.data = append(t.data, data[pos:end]...)
		pos = end
	}

	for i, t := range cs.tiles {
		if t == nil {
			return nil, errFormat("missing tile-part for a tile")
		}
		_ = i
	}
	return cs, nil
}

func parseSIZ(b []byte, s *sizInfo) error {
	if len(b) < 36 {
		return errFormat("SIZ too short")
	}
	s.xsiz, s.ysiz = be32(b[2:]), be32(b[6:])
	s.xosiz, s.yosiz = be32(b[10:]), be32(b[14:])
	s.xtsiz, s.ytsiz = be32(b[18:]), be32(b[22:])
	s.xtosiz, s.ytosiz = be32(b[26:]), be32(b[30:])
	n := be16(b[34:])
	if n < 1 || n > 16384 {
		return errFormat("bad component count")
	}
	if len(b) < 36+3*n {
		return errFormat("SIZ too short for components")
	}
	if int64(s.xsiz-s.xosiz)*int64(s.ysiz-s.yosiz) > int64(MaxImagePixels) {
		return errFormat("image exceeds MaxImagePixels")
	}
	if s.xsiz <= s.xosiz || s.ysiz <= s.yosiz ||
		s.xtsiz <= 0 || s.ytsiz <= 0 ||
		s.xtosiz > s.xosiz || s.ytosiz > s.yosiz ||
		s.xtosiz+s.xtsiz <= s.xosiz || s.ytosiz+s.ytsiz <= s.yosiz {
		return errFormat("inconsistent SIZ geometry")
	}
	s.comps = make([]sizComponent, n)
	for i := 0; i < n; i++ {
		ss := b[36+3*i]
		s.comps[i] = sizComponent{
			depth:  int(ss&0x7F) + 1,
			signed: ss&0x80 != 0,
			dx:     int(b[37+3*i]),
			dy:     int(b[38+3*i]),
		}
		c := &s.comps[i]
		if c.depth > 38 || c.dx < 1 || c.dy < 1 {
			return errFormat("bad component parameters")
		}
	}
	return nil
}

// parseSPcox parses the SPcod/SPcoc tail shared by COD and COC.
func parseSPcox(b []byte, precincts bool, cp *codingParams) error {
	if len(b) < 5 {
		return errFormat("coding style segment too short")
	}
	cp.numDecomps = int(b[0])
	if cp.numDecomps > 32 {
		return errFormat("too many decomposition levels")
	}
	cp.cbW = uint(b[1]&0x0F) + 2
	cp.cbH = uint(b[2]&0x0F) + 2
	if cp.cbW > 10 || cp.cbH > 10 || cp.cbW+cp.cbH > 12 {
		return errFormat("bad code-block size")
	}
	cp.cbStyle = b[3]
	cp.transform = b[4]
	if cp.transform > 1 {
		return errFormat("unknown wavelet transform")
	}
	nRes := cp.numDecomps + 1
	cp.precW = make([]uint, nRes)
	cp.precH = make([]uint, nRes)
	if precincts {
		if len(b) < 5+nRes {
			return errFormat("truncated precinct sizes")
		}
		for r := 0; r < nRes; r++ {
			cp.precW[r] = uint(b[5+r] & 0x0F)
			cp.precH[r] = uint(b[5+r] >> 4)
			if r > 0 && (cp.precW[r] == 0 || cp.precH[r] == 0) {
				return errFormat("invalid precinct size")
			}
		}
	} else {
		for r := 0; r < nRes; r++ {
			cp.precW[r], cp.precH[r] = 15, 15
		}
	}
	return nil
}

func parseCOD(b []byte, h *codHeader) error {
	if len(b) < 10 {
		return errFormat("COD too short")
	}
	scod := b[0]
	h.sop = scod&0x02 != 0
	h.eph = scod&0x04 != 0
	h.progression = b[1]
	if h.progression > progCPRL {
		return errFormat("unknown progression order")
	}
	h.numLayers = be16(b[2:])
	if h.numLayers < 1 {
		return errFormat("bad layer count")
	}
	h.mct = b[4] != 0
	return parseSPcox(b[5:], scod&0x01 != 0, &h.cp)
}

func parseCOC(b []byte, numComps int, base *codingParams) (int, *codingParams, error) {
	i := 0
	var comp int
	if numComps < 257 {
		if len(b) < 2 {
			return 0, nil, errFormat("COC too short")
		}
		comp = int(b[0])
		i = 1
	} else {
		if len(b) < 3 {
			return 0, nil, errFormat("COC too short")
		}
		comp = be16(b)
		i = 2
	}
	if comp >= numComps {
		return 0, nil, errFormat("COC component out of range")
	}
	scoc := b[i]
	cp := &codingParams{}
	if err := parseSPcox(b[i+1:], scoc&0x01 != 0, cp); err != nil {
		return 0, nil, err
	}
	return comp, cp, nil
}

func parseQuantBody(b []byte, qp *quantParams) error {
	if len(b) < 1 {
		return errFormat("quantization segment too short")
	}
	sq := b[0]
	qp.style = int(sq & 0x1F)
	qp.guardBits = int(sq >> 5)
	body := b[1:]
	switch qp.style {
	case 0:
		qp.exps = make([]int, len(body))
		for i, v := range body {
			qp.exps[i] = int(v >> 3)
		}
	case 1:
		if len(body) < 2 {
			return errFormat("truncated quantization values")
		}
		v := be16(body)
		qp.exps = []int{v >> 11}
		qp.mants = []int{v & 0x7FF}
	case 2:
		if len(body)%2 != 0 {
			return errFormat("truncated quantization values")
		}
		n := len(body) / 2
		qp.exps = make([]int, n)
		qp.mants = make([]int, n)
		for i := 0; i < n; i++ {
			v := be16(body[2*i:])
			qp.exps[i] = v >> 11
			qp.mants[i] = v & 0x7FF
		}
	default:
		return errFormat("unknown quantization style")
	}
	return nil
}

func parseQCD(b []byte, qp *quantParams) error {
	return parseQuantBody(b, qp)
}

func parseQCC(b []byte, numComps int) (int, *quantParams, error) {
	var comp, i int
	if numComps < 257 {
		if len(b) < 1 {
			return 0, nil, errFormat("QCC too short")
		}
		comp, i = int(b[0]), 1
	} else {
		if len(b) < 2 {
			return 0, nil, errFormat("QCC too short")
		}
		comp, i = be16(b), 2
	}
	if comp >= numComps {
		return 0, nil, errFormat("QCC component out of range")
	}
	qp := &quantParams{}
	if err := parseQuantBody(b[i:], qp); err != nil {
		return 0, nil, err
	}
	return comp, qp, nil
}
