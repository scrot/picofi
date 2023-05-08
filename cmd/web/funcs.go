package main

import (
	"fmt"
	"html/template"

	"github.com/Rhymond/go-money"
	"github.com/scrot/picofi"
)

type TF struct {
	calculator *picofi.Calculator
}

func NewTF(c *picofi.Calculator) TF {
	return TF{c}
}

func (tf TF) FuncMap() template.FuncMap {
	return template.FuncMap{
		"annualSaveRate":           tf.annualSaveRate,
		"annualSaveRatePercentage": tf.annualSaveRatePercentage,
	}

}

func (tf TF) annualSaveRate(income, expenses float64) string {
	res, err := tf.calculator.AnnualSaveRate(
		money.NewFromFloat(income, tf.calculator.Currency),
		money.NewFromFloat(expenses, tf.calculator.Currency),
	)

	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	return res.Display()
}

func (tf TF) annualSaveRatePercentage(income, expenses float64) string {
	res, err := tf.calculator.AnnualSaveRatePercentage(
		money.NewFromFloat(income, tf.calculator.Currency),
		money.NewFromFloat(expenses, tf.calculator.Currency),
	)

	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	return fmt.Sprintf("%.2f%%", res)
}
