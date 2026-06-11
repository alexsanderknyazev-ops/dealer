package service

import "testing"

func TestMultiplyAmount(t *testing.T) {
	if got := multiplyAmount("2", "850"); got != "1700.00" {
		t.Fatalf("got %q", got)
	}
	if got := multiplyAmount("1.5", "2000"); got != "3000.00" {
		t.Fatalf("got %q", got)
	}
}

func TestSumAmounts(t *testing.T) {
	if got := sumAmounts("2000.00", "1700.00"); got != "3700.00" {
		t.Fatalf("got %q", got)
	}
}

func TestQuantityToUnits(t *testing.T) {
	q, err := quantityToUnits("2")
	if err != nil || q != 2 {
		t.Fatalf("err=%v q=%d", err, q)
	}
	q, err = quantityToUnits("2.1")
	if err != nil || q != 3 {
		t.Fatalf("ceil err=%v q=%d", err, q)
	}
	if _, err = quantityToUnits("0"); err == nil {
		t.Fatal("want error for zero quantity")
	}
}
