package picofi

import (
	"errors"
	"fmt"

	"github.com/Rhymond/go-money"
	"golang.org/x/exp/slog"
)

var (
	ErrNegativeIncome = errors.New("income can't be negative, don't include debt here")
	ErrInvalidSavings = errors.New("savings cannot be higer than the income")
)

type Calculator struct {
	Logger   *slog.Logger
	Currency string
}

func NewCalculator(logger *slog.Logger, currency money.Currency) Calculator {
	return Calculator{logger, currency.Code}
}

// AnnualSaveRate calculates the annual expenses using income - expenses
// a negative save rate denotes a debt, saving more that is earned throws an ErrInvalidSavings
// a negative income throws a ErrNegativeIncome, don't combine debt with income
func (c Calculator) AnnualSaveRate(income *money.Money, expenses *money.Money) (*money.Money, error) {
	if income.Amount() < 0 {
		return nil, ErrNegativeIncome
	}
	savings, err := income.Subtract(expenses)
	if err != nil {
		c.Logger.Error(fmt.Sprintf("annualSaveRate:%s - %s", income.Display(), expenses.Display()), "err", err)
		return nil, err
	}

	gt, err := savings.GreaterThan(income)
	if err != nil {
		return nil, err
	}

	if gt {
		return nil, ErrInvalidSavings
	}

	return savings, nil
}

// AnnualSaveRatePercentage caclulates the percentages of money saved annually based on income and
// expenses. It uses AnnualSaveRate to calculate the savings.
func (c Calculator) AnnualSaveRatePercentage(income *money.Money, expenses *money.Money) (float64, error) {
	savings, err := c.AnnualSaveRate(income, expenses)
	if err != nil {
		return 0, err
	}

	sp := savings.AsMajorUnits() / income.AsMajorUnits() * 100

	return sp, nil
}
