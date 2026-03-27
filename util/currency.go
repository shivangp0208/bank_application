package util

const (
	USD string = "USD"
	CAD string = "CAD"
	EUR string = "EUR"
	INR string = "INR"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, CAD, EUR, INR:
		return true
	}
	return false
}
