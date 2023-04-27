package picofi

import (
	"fmt"

	"github.com/Rhymond/go-money"
	"golang.org/x/exp/slog"
)

type Calculator struct {
	Logger   *slog.Logger
	Currency string
}

func NewCalculator(logger *slog.Logger, currency money.Currency) Calculator {
	return Calculator{logger, currency.Code}
}

func (c Calculator) AnnualSaveRate(income *money.Money, expenses *money.Money) *money.Money {
	result, err := income.Subtract(expenses)
	if err != nil {
		c.Logger.Error(fmt.Sprintf("annualSaveRate:%s - %s", income.Display(), expenses.Display()), "err", err)
	}
	return result
}
