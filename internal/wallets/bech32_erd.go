package wallets

import (
	"fmt"
)

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func PubKeyToBech32(pubKey32 []byte) (string, error) {
	return bech32EncodeErdFromPubKey(pubKey32)
}

func bech32EncodeErdFromPubKey(pubKey32 []byte) (string, error) {
	if len(pubKey32) != 32 {
		return "", fmt.Errorf("pubkey must be 32 bytes, got %d", len(pubKey32))
	}

	data5, err := convertBits(pubKey32, 8, 5, true)
	if err != nil {
		return "", err
	}

	return bech32Encode("erd", data5)
}

func bech32Encode(hrp string, data []byte) (string, error) {
	checksum := bech32CreateChecksum(hrp, data)
	combined := append(data, checksum...)

	ret := hrp + "1"
	for _, p := range combined {
		if int(p) >= len(bech32Charset) {
			return "", fmt.Errorf("bech32: invalid data value %d", p)
		}
		ret += string(bech32Charset[p])
	}
	return ret, nil
}

func bech32HrpExpand(hrp string) []byte {
	expand := make([]byte, 0, len(hrp)*2+1)
	for i := 0; i < len(hrp); i++ {
		expand = append(expand, hrp[i]>>5)
	}
	expand = append(expand, 0)
	for i := 0; i < len(hrp); i++ {
		expand = append(expand, hrp[i]&31)
	}
	return expand
}

func bech32Polymod(values []byte) uint32 {
	var chk uint32 = 1
	generator := [5]uint32{
		0x3b6a57b2,
		0x26508e6d,
		0x1ea119fa,
		0x3d4233dd,
		0x2a1462b3,
	}

	for _, v := range values {
		top := chk >> 25
		chk = (chk & 0x1ffffff) << 5
		chk ^= uint32(v)
		for i := 0; i < 5; i++ {
			if ((top >> uint(i)) & 1) == 1 {
				chk ^= generator[i]
			}
		}
	}
	return chk
}

func bech32CreateChecksum(hrp string, data []byte) []byte {
	values := append(bech32HrpExpand(hrp), data...)
	values = append(values, 0, 0, 0, 0, 0, 0)
	mod := bech32Polymod(values) ^ 1

	ret := make([]byte, 6)
	for i := 0; i < 6; i++ {
		ret[i] = byte((mod >> uint(5*(5-i))) & 31)
	}
	return ret
}

func convertBits(data []byte, fromBits, toBits uint, pad bool) ([]byte, error) {
	var ret []byte
	var acc uint
	var bits uint
	maxv := uint((1 << toBits) - 1)
	maxAcc := uint((1 << (fromBits + toBits - 1)) - 1)

	for _, value := range data {
		if (uint(value) >> fromBits) != 0 {
			return nil, fmt.Errorf("convertBits: invalid data range: %d", value)
		}
		acc = ((acc << fromBits) | uint(value)) & maxAcc
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			ret = append(ret, byte((acc>>bits)&maxv))
		}
	}

	if pad {
		if bits > 0 {
			ret = append(ret, byte((acc<<(toBits-bits))&maxv))
		}
	} else {
		if bits >= fromBits {
			return nil, fmt.Errorf("convertBits: illegal zero padding")
		}
		if ((acc << (toBits - bits)) & maxv) != 0 {
			return nil, fmt.Errorf("convertBits: non-zero padding")
		}
	}

	return ret, nil
}
