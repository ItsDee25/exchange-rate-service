package constants

var SupportedCurrencies = map[string]struct{}{
	"USD": {}, "INR": {}, "EUR": {}, "JPY": {}, "GBP": {},
}

var SupportedCurrencyPairs = [][2]string{
	{"USD", "INR"},
	{"USD", "EUR"},
	{"USD", "GBP"},
	{"USD", "JPY"},
	{"USD", "CAD"},
	{"USD", "AUD"},
	{"EUR", "INR"},
	{"EUR", "USD"},
	{"EUR", "GBP"},
	{"EUR", "JPY"},
	{"GBP", "USD"},
	{"GBP", "INR"},
	{"GBP", "EUR"},
	{"INR", "USD"},
	{"INR", "EUR"},
	{"INR", "GBP"},
	{"JPY", "USD"},
	{"JPY", "EUR"},
	{"CAD", "USD"},
	{"AUD", "USD"},
}

