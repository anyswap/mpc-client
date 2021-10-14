package mpcrpc

import (
	"errors"
	"math/big"
	"strconv"
	"time"
)

// NowMilliStr returns now timestamp in miliseconds of string format.
func NowMilliStr() string {
	return strconv.FormatInt((time.Now().UnixNano() / 1e6), 10)
}

// GetBigIntFromStr new big int from string.
func GetBigIntFromStr(str string) (*big.Int, error) {
	bi, ok := new(big.Int).SetString(str, 0)
	if !ok || bi.BitLen() > 256 {
		return nil, errors.New("invalid 256 bit integer: " + str)
	}
	return bi, nil
}
