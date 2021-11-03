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

// GetUint64FromStr get uint64 from string.
func GetUint64FromStr(str string) (uint64, error) {
	res, ok := ParseUint64(str)
	if !ok {
		return 0, errors.New("invalid unsigned 64 bit integer: " + str)
	}
	return res, nil
}

// ParseUint64 parses s as an integer in decimal or hexadecimal syntax.
// Leading zeros are accepted. The empty string parses as zero.
func ParseUint64(s string) (uint64, bool) {
	if s == "" {
		return 0, true
	}
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		v, err := strconv.ParseUint(s[2:], 16, 64)
		return v, err == nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	return v, err == nil
}
