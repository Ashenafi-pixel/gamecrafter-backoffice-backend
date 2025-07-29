package utils

import (
	"fmt"
	"math/rand"
	mathRand "math/rand"
	"strings"
	"time"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/shopspring/decimal"
)

func GenerateRandomCrashPoint(min, max int64) (decimal.Decimal, error) {
	if min >= max {
		return decimal.Zero, fmt.Errorf("min must be less than max")
	}
	randFloat := mathRand.Float64()
	diff := float64(max-min) + 1
	randValue := float64(min) + randFloat*(diff)
	randDecimal := decimal.NewFromFloat(randValue)
	if randDecimal.GreaterThan(decimal.NewFromInt(max)) {
		randDecimal = decimal.NewFromInt(max)
	}

	weightDecay := randFloat

	scaledValue := randDecimal.Mul(decimal.NewFromFloat(weightDecay))

	return scaledValue, nil
}

func GenerateTransactionId() string {
	timeStamp := time.Now().Unix()
	mathRand.Seed(timeStamp)
	transactionId := mathRand.Intn(100000000)
	return fmt.Sprintf("%09d", transactionId)
}

func GenerateRandomUsername(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	mathRand.Seed(time.Now().Unix())
	var sb strings.Builder
	for i := 0; i < length; i++ {
		index := mathRand.Intn(len(chars))
		sb.WriteByte(chars[index])
	}
	return sb.String()
}

func GenerateRandomPassowrd(length int) string {
	lowerCase := "abcdefghijklmnopqrstuvwxyz."
	upperCase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	specialChars := "@#$%^&*"
	allChars := lowerCase + upperCase + digits + specialChars
	mathRand.Seed(time.Now().Unix())
	var password strings.Builder
	password.WriteByte(lowerCase[mathRand.Intn(len(lowerCase))])
	password.WriteByte(upperCase[mathRand.Intn(len(upperCase))])
	password.WriteByte(digits[mathRand.Intn(len(digits))])
	password.WriteByte(specialChars[mathRand.Intn(len(specialChars))])
	for i := 4; i < length; i++ {
		password.WriteByte(allChars[mathRand.Intn(len(allChars))])
	}
	return password.String()
}

func RandomDecimal(min, max int) float64 {
	rand.Seed(time.Now().UnixNano())
	rangeSize := float64(max - min)
	return float64(min) + (rand.Float64() * rangeSize)
}
func GetPrefix(val dto.Type) string {
	switch val {
	case dto.INFLUENCER:
		return "IN-"
	case dto.OTG_AGENT:
		return "OTG-"
	case dto.PLAYER:
		return "P-"
	default:
		return ""
	}
}
