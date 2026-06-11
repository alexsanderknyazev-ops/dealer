package service

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func multiplyAmount(qty, price string) string {
	q, _ := strconv.ParseFloat(strings.TrimSpace(qty), 64)
	p, _ := strconv.ParseFloat(strings.TrimSpace(price), 64)
	return fmt.Sprintf("%.2f", q*p)
}

func sumAmounts(amounts ...string) string {
	var total float64
	for _, a := range amounts {
		v, _ := strconv.ParseFloat(strings.TrimSpace(a), 64)
		total += v
	}
	return fmt.Sprintf("%.2f", total)
}

func quantityToUnits(qty string) (int32, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(qty), 64)
	if err != nil || f <= 0 {
		return 0, fmt.Errorf("invalid quantity")
	}
	return int32(math.Ceil(f - 1e-9)), nil
}
