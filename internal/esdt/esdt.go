package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func TokenIDHex(tokenID string) string {
	return hex.EncodeToString([]byte(tokenID))
}

func BigIntHexEven(x *big.Int) string {
	if x.Sign() < 0 {
		panic("negative amount not allowed")
	}
	s := strings.TrimLeft(x.Text(16), "0")
	if s == "" {
		s = "0"
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return strings.ToLower(s)
}

func AmountToBaseUnits(amount string, decimals int) (*big.Int, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return nil, fmt.Errorf("empty amount")
	}
	if strings.HasPrefix(amount, "-") {
		return nil, fmt.Errorf("negative amount")
	}

	parts := strings.SplitN(amount, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	if intPart == "" {
		intPart = "0"
	}
	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}

	if len(fracPart) > decimals {
		return nil, fmt.Errorf("too many decimal places: got %d, max %d", len(fracPart), decimals)
	}
	fracPart = fracPart + strings.Repeat("0", decimals-len(fracPart))

	base := new(big.Int)
	if _, ok := base.SetString(intPart, 10); !ok {
		return nil, fmt.Errorf("invalid integer part")
	}

	mul := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	base.Mul(base, mul)

	frac := new(big.Int)
	if fracPart == "" {
		frac.SetInt64(0)
	} else if _, ok := frac.SetString(fracPart, 10); !ok {
		return nil, fmt.Errorf("invalid fractional part")
	}

	base.Add(base, frac)
	return base, nil
}

func ESDTTransferData(tokenID string, amountBaseUnits *big.Int) string {
	return "ESDTTransfer@" + TokenIDHex(tokenID) + "@" + BigIntHexEven(amountBaseUnits)
}
