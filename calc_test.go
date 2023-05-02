package picofi

import (
	"testing"

	"github.com/Rhymond/go-money"
	"golang.org/x/exp/slog"
)

func TestAnnualSavings(t *testing.T) {
	tests := []struct {
		name     string
		income   float64
		expenses float64
		expected *money.Money
	}{
		{"noSavings", 100000, 100000, money.NewFromFloat(0.0, money.EUR)},
		{"halfSavings", 100000, 50000, money.NewFromFloat(50000.0, money.EUR)},
		{"negSavings", 50000, 100000, money.NewFromFloat(-50000.0, money.EUR)},
	}

	c := NewCalculator(slog.Default(), *money.GetCurrency(money.EUR))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i := money.NewFromFloat(test.income, money.EUR)
			e := money.NewFromFloat(test.expenses, money.EUR)
			r := c.AnnualSaveRate(i, e)
			eq, err := r.Equals(test.expected)
			if err != nil {
				t.Errorf("unable to compare %s with %s", test.expected.Display(), r.Display())
			}

			if !eq {
				t.Errorf("expected savings of %s but got %s", test.expected.Display(), r.Display())
			}
		})
	}
}

func FuzzAnnualSavings(f *testing.F) {
	c := NewCalculator(slog.Default(), *money.GetCurrency(money.EUR))

	f.Fuzz(func(t *testing.T, income, expenses float64) {
		i := money.NewFromFloat(income, money.EUR)
		e := money.NewFromFloat(expenses, money.EUR)
		r := c.AnnualSaveRate(i, e)

		cmp, err := e.Add(r)
		if err != nil {
			t.Errorf("unable to add %s to %s", e.Display(), r.Display())
		}

		eq, err := cmp.Equals(i)
		if err != nil {
			t.Errorf("unable to compare %s to %s", cmp.Display(), i.Display())
		}

		if !eq {
			t.Errorf("adding expenses %s and savings %s should be equal to income %s but got %s", e.Display(), r.Display(), i.Display(), cmp.Display())
		}
	})

}
