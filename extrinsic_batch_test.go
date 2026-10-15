package substrate_test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/itering/scale.go/types"
	substrate "github.com/itering/substrate-api-rpc"
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/stretchr/testify/require"
)

func batchMetadata() *metadata.Instant {
	return &metadata.Instant{
		MetadataVersion: 13,
		CallIndex: map[string]types.CallIndex{"0000": {
			Module: types.MetadataModules{Name: "Example"},
			Call: types.MetadataCalls{Name: "set_value", Args: []types.MetadataModuleCallArgument{
				{Name: "value", Type: "U32"},
			}},
		}},
	}
}

func batchExtrinsic(value uint32) string {
	// Compact body length 7; bare V4; call 0000; one U32 argument.
	raw := []byte{28, 4, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(raw[4:], value)
	return "0x" + hex.EncodeToString(raw)
}

func TestDecodeExtrinsicBatchFailure(t *testing.T) {
	for _, invalid := range []struct{ name, raw, cause string }{
		{"unknown call", "0x0c04ffff", "Not find Extrinsic Lookup"},
		{"invalid version", "0x0c060000", "version 06"},
		{"invalid hex", "0x1c040000zz000000", "invalid byte"},
		{"odd hex", "0x1c040000010000000", "odd length"},
		{"trailing byte", "0x2004000001000000ff", "consumed"},
	} {
		for _, index := range []int{0, 1, 2} {
			t.Run(fmt.Sprintf("%s/index-%d", invalid.name, index), func(t *testing.T) {
				raw := []string{batchExtrinsic(11), batchExtrinsic(22), batchExtrinsic(33)}
				raw[index] = invalid.raw
				decoded, err := substrate.DecodeExtrinsic(raw, batchMetadata(), 1)
				require.Error(t, err)
				require.Nil(t, decoded, "discard every decoded prefix on failure")
				require.Contains(t, err.Error(), fmt.Sprintf("index %d", index))
				require.Contains(t, err.Error(), invalid.cause)
			})
		}
	}
}

func TestDecodeExtrinsicBatchSuccess(t *testing.T) {
	values := []uint32{11, 22, 33}
	raw := make([]string, len(values))
	for index, value := range values {
		raw[index] = batchExtrinsic(value)
	}
	decoded, err := substrate.DecodeExtrinsic(raw, batchMetadata(), 1)
	require.NoError(t, err)
	require.Len(t, decoded, len(raw))
	for index, result := range decoded {
		require.Equal(t, "set_value", result["call_module_function"])
		params := result["params"].([]interface{})
		require.Len(t, params, 1)
		require.Equal(t, float64(values[index]), params[0].(map[string]interface{})["value"])
	}
	// A failed batch must not poison a corrected retry or reorder its results.
	failed, err := substrate.DecodeExtrinsic([]string{raw[0], "0x0c04ffff", raw[2]}, batchMetadata(), 1)
	require.Error(t, err)
	require.Nil(t, failed)
	retried, err := substrate.DecodeExtrinsic(raw, batchMetadata(), 1)
	require.NoError(t, err)
	require.Equal(t, decoded, retried)
}

func TestDecodeExtrinsicEmptyAndMissingMetadata(t *testing.T) {
	for _, raw := range [][]string{nil, {}} {
		decoded, err := substrate.DecodeExtrinsic(raw, batchMetadata(), 1)
		require.NoError(t, err)
		require.Empty(t, decoded)
	}
	decoded, err := substrate.DecodeExtrinsic([]string{batchExtrinsic(1)}, nil, 1)
	require.Nil(t, decoded)
	require.EqualError(t, err, "decode extrinsic batch: metadata is nil")
}
