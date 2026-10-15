package substrate

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	scale "github.com/itering/scale.go"
	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/itering/substrate-api-rpc/util"
	"github.com/shopspring/decimal"
)

// DecodeExtrinsic decodes a complete batch in input order. Any failure returns
// nil results; per-extrinsic errors identify the original zero-based index.
func DecodeExtrinsic(rawList []string, metadata *metadata.Instant, spec int) (r []map[string]interface{}, err error) {
	if metadata == nil {
		return nil, fmt.Errorf("decode extrinsic batch: metadata is nil")
	}
	index := 0
	defer func() {
		if fatal := recover(); fatal != nil {
			r = nil
			if cause, ok := fatal.(error); ok {
				err = fmt.Errorf("decode extrinsic at index %d: %w", index, cause)
			} else {
				err = fmt.Errorf("decode extrinsic at index %d: %v", index, fatal)
			}
		}
	}()
	m := types.MetadataStruct(*metadata)
	for index = range rawList {
		raw, decodeErr := hex.DecodeString(strings.TrimPrefix(rawList[index], "0x"))
		if decodeErr != nil {
			return nil, fmt.Errorf("decode extrinsic at index %d: invalid hex: %w", index, decodeErr)
		}
		e := scale.ExtrinsicDecoder{}
		option := types.ScaleDecoderOption{Metadata: &m, Spec: spec}
		e.Init(scaleBytes.ScaleBytes{Data: raw}, &option)
		e.Process()
		if e.Data.Offset != len(raw) {
			return nil, fmt.Errorf("decode extrinsic at index %d: consumed %d of %d bytes", index, e.Data.Offset, len(raw))
		}
		decoded := e.Value.(*scale.GenericExtrinsic).ToMap()
		if decoded == nil {
			return nil, fmt.Errorf("decode extrinsic at index %d: result conversion failed", index)
		}
		r = append(r, decoded)
	}
	return
}

type Mortal struct {
	Period uint64
	Phase  uint64
}

func DecodeMortal(era string) *Mortal {
	if era == "" || era == "00" {
		return nil
	}
	eraU8a := util.HexToBytes(era)
	first := uint64(eraU8a[0])
	second := uint64(eraU8a[1])
	encoded := first + (second << 8)
	var period uint64 = 2 << (encoded % (1 << 4))
	quantizeFactor := math.Max(float64(period>>12), 1)
	phase := (encoded >> 4) * uint64(quantizeFactor)
	return &Mortal{
		Period: period,
		Phase:  phase,
	}
}

func (m *Mortal) Birth(current uint64) uint64 {
	s := (decimal.Max(decimal.New(int64(current), 0), decimal.New(int64(m.Phase), 0)).
		Sub(decimal.New(int64(m.Phase), 0))).Div(decimal.New(int64(m.Period), 0)).IntPart()
	return uint64(s)*m.Period + m.Phase
}

func (m *Mortal) Death(current uint64) uint64 {
	return m.Birth(current) + m.Period
}
