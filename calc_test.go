package picofi

import (
	"errors"
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
			r, err := c.AnnualSaveRate(i, e)
			if err != nil {
				t.Errorf("unable to calculate save rate: %s", err)
			}
			eq, err := r.Equals(test.expected)
			if err != nil {
				t.Errorf("unable to compare %s with %s: %s", test.expected.Display(), r.Display(), err)
			}

			if !eq {
				t.Errorf("expected savings of %s but got %s", test.expected.Display(), r.Display())
			}
		})
	}
}

func TestAnnualSavingsPercentage(t *testing.T) {
	tests := []struct {
		name          string
		income        float64
		expenses      float64
		expected      float64
		expectedError error
	}{
		{"positiveSavings", 100000, 20000, 80.0, nil},              // percentage of income saved
		{"zeroSavings", 100000, 100000, 0.0, nil},                  // nothing of income saved
		{"invalidSavings", 100000, -10000, 0.0, ErrInvalidSavings}, // err savings more than income
		{"invalidIncome", -100000, 10000, 0.0, ErrNegativeIncome},  // err negative income
		{"negativeSavings", 100000, 150000, -50.0, nil},            // no savings equals debt
	}

	c := NewCalculator(slog.Default(), *money.GetCurrency(money.EUR))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i := money.NewFromFloat(test.income, money.EUR)
			e := money.NewFromFloat(test.expenses, money.EUR)
			r, err := c.AnnualSaveRatePercentage(i, e)
			if err != nil && !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %s", err)
			}

			if r != test.expected {
				t.Errorf("expected savings of %f%% but got %f%%", test.expected, r)
			}
		})
	}

}

// func FuzzAnnualSavings(f *testing.F) {
// 	c := NewCalculator(slog.Default(), *money.GetCurrency(money.EUR))

// 	f.Fuzz(func(t *testing.T, income, expenses float64) {
// 		i := money.NewFromFloat(income, money.EUR)
// 		e := money.NewFromFloat(expenses, money.EUR)
// 		r, err := c.AnnualSaveRate(i, e)

// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}

// 		cmp, err := e.Add(r)
// 		if err != nil {
// 			t.Errorf("unable to add %s to %s", e.Display(), r.Display())
// 		}

// 		eq, err := cmp.Equals(i)
// 		if err != nil {
// 			t.Errorf("unable to compare %s to %s", cmp.Display(), i.Display())
// 		}

// 		if !eq {
// 			t.Errorf("adding expenses %s and savings %s should be equal to income %s but got %s", e.Display(), r.Display(), i.Display(), cmp.Display())
// 		}
// 	})

// }
