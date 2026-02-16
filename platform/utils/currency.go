package utils

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// FormatCurrency formats a decimal amount as currency with proper formatting
func FormatCurrency(amount decimal.Decimal, currency string) string {
	// Convert to float64 for formatting
	amountFloat, _ := amount.Float64()
	
	// Format with commas and 2 decimal places
	formatted := fmt.Sprintf("%.2f", amountFloat)
	
	// Add commas for thousands separator
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]
	decimalPart := parts[1]
	
	// Add commas to integer part
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}
	
	// Combine with currency symbol
	switch strings.ToUpper(currency) {
	case "USD":
		return fmt.Sprintf("$%s.%s", integerPart, decimalPart)
	case "EUR":
		return fmt.Sprintf("€%s.%s", integerPart, decimalPart)
	case "GBP":
		return fmt.Sprintf("£%s.%s", integerPart, decimalPart)
	case "NGN":
		return fmt.Sprintf("₦%s.%s", integerPart, decimalPart)
	case "P":
		return fmt.Sprintf("P%s.%s", integerPart, decimalPart)
	default:
		return fmt.Sprintf("%s %s.%s", currency, integerPart, decimalPart)
	}
}

// FormatNumber formats a decimal amount as a number with commas (no currency symbol)
func FormatNumber(amount decimal.Decimal) string {
	// Convert to float64 for formatting
	amountFloat, _ := amount.Float64()
	
	// Format with commas and 2 decimal places
	formatted := fmt.Sprintf("%.2f", amountFloat)
	
	// Add commas for thousands separator
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]
	decimalPart := parts[1]
	
	// Add commas to integer part
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}
	
	return fmt.Sprintf("%s.%s", integerPart, decimalPart)
}

// FormatCurrencyShort formats currency for mobile/short display
func FormatCurrencyShort(amount decimal.Decimal, currency string) string {
	amountFloat, _ := amount.Float64()
	
	switch {
	case amountFloat >= 1000000:
		shortAmount := amountFloat / 1000000
		return fmt.Sprintf("%s%.1fM", getCurrencySymbol(currency), shortAmount)
	case amountFloat >= 1000:
		shortAmount := amountFloat / 1000
		return fmt.Sprintf("%s%.1fK", getCurrencySymbol(currency), shortAmount)
	default:
		return FormatCurrency(amount, currency)
	}
}

// getCurrencySymbol returns the currency symbol for a given currency code
func getCurrencySymbol(currency string) string {
	switch strings.ToUpper(currency) {
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "NGN":
		return "₦"
	case "P":
		return "P"
	default:
		return currency + " "
	}
}