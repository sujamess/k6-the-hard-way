package uniquer_test

import (
	"testing"

	"github.com/sujamess/k6-the-hard-way/pkgs/uniquer"
)

func TestOrderNumber(t *testing.T) {
	orderNumber := uniquer.OrderNumber()
	if len(orderNumber) != 16 {
		t.Errorf("want length %d, got %d", 16, len(orderNumber))
	}
}

func BenchmarkOrderNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		uniquer.OrderNumber()
	}
}
