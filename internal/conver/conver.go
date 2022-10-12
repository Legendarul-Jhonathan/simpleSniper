package convert

import (
	"encoding/hex"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
)

func HexToUint64(hexStr string) uint64 {
	// remove 0x suffix if found in the input string
	cleaned := strings.Replace(hexStr, "0x", "", -1)

	// base 16 for hexadecimal
	result, _ := strconv.ParseUint(cleaned, 16, 64)
	return uint64(result)
}

//HexToString transforms hex to string
func HexToString(hexString string) string {

	hexString = strings.Replace(hexString, "0x", "", -1)

	decode, err := hex.DecodeString(hexString)
	if err != nil {
		log.Fatal(err)
	}

	return string(decode)
}

//HexToStringRemoveNulls transforms hex to string
func HexToStringRemoveNulls(hexString string) string {

	hexString = strings.Replace(hexString, "0x", "", -1)
	hexString = removeNulls(hexString)

	decode, err := hex.DecodeString(hexString)
	if err != nil {
		log.Fatal(err)
	}

	return string(decode)
}

func removeNulls(hexString string) string {

	if len(hexString) != 192 {
		hexString = strings.Replace(hexString, "0x", "", -1)
		hexString = strings.Replace(hexString, "00", "", -1)
		return hexString
	}

	hexString = strings.Replace(hexString, "0x", "", -1)
	hexString = strings.Replace(hexString, "0000000000000000000000000000000000000000000000000000000000000020", "", -1)

	for i := 0; i < len(hexString); i = i + 4 {
		chunk := hexString[i : i+4]
		if chunk != "0000" {
			lenght := HexToInt32(chunk)
			if int(lenght) > len(hexString) {
				break
			}
			hexString = hexString[i+4 : i+4+(int(lenght)*2)]
			break
		}
	}

	return hexString
}

//WeiToEth transforms wei to ETH
func WeiToEth(wei *big.Int) *big.Float {
	c := new(big.Int)
	c.SetString("1000000000000000000", 0)

	var val *big.Float = new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt(c))
	return val
}

func SetDecimal(number big.Int, decilams int) big.Float {
	c := new(big.Int)
	//TODO: why do I have to add +1 ???
	c.SetString(StrPad("1", decilams+1, "0", "RIGHT"), 0)
	var val *big.Float = new(big.Float).Quo(new(big.Float).SetInt(&number), new(big.Float).SetInt(c))
	return *val
}

//HexToBigInt transform a hex to bigInt
func HexToBigInt(hexaString string) *big.Int {
	// replace 0x or 0X with empty String
	numberStr := strings.Replace(hexaString, "0x", "", -1)
	n := new(big.Int)
	n.SetString(numberStr, 16)

	return n
}

//HexToInt64 transform a hex to int64
func HexToInt64(hexaString string) int64 {
	// replace 0x or 0X with empty String
	numberStr := strings.Replace(hexaString, "0x", "", -1)
	n, _ := strconv.ParseInt(numberStr, 16, 64)

	return n
}

//HexToInt32 transform a hex to int64
func HexToInt32(hexaString string) int32 {
	// replace 0x or 0X with empty String
	numberStr := strings.Replace(hexaString, "0x", "", -1)
	n, _ := strconv.ParseInt(numberStr, 16, 32)

	return int32(n)
}

//Int64ToHex transforms a int to hex
func Int64ToHex(number int64) string {
	return "0x" + strconv.FormatInt(number, 16)
}
func StrPad(input string, padLength int, padString string, padType string) string {
	var output string

	inputLength := len(input)
	padStringLength := len(padString)

	if inputLength >= padLength {
		return input
	}

	repeat := math.Ceil(float64(1) + (float64(padLength-padStringLength))/float64(padStringLength))

	switch padType {
	case "RIGHT":
		output = input + strings.Repeat(padString, int(repeat))
		output = output[:padLength]
	case "LEFT":
		output = strings.Repeat(padString, int(repeat)) + input
		output = output[len(output)-padLength:]
	case "BOTH":
		length := (float64(padLength - inputLength)) / float64(2)
		repeat = math.Ceil(length / float64(padStringLength))
		output = strings.Repeat(padString, int(repeat))[:int(math.Floor(float64(length)))] + input + strings.Repeat(padString, int(repeat))[:int(math.Ceil(float64(length)))]
	}

	return output
}
