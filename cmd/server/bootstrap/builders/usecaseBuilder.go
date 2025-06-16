package builders

import usecase "github.com/ItsDee25/exchange-rate-service/internal/usecase/currency"

type Usecases struct {
	CurrencyUsecase *usecase.CurrencyUsecase
}

func NewUsecases() *Usecases {
	return &Usecases{}
}

func (u *Usecases) WithCurrencyUsecase(c *usecase.CurrencyUsecase) *Usecases {
	u.CurrencyUsecase = c
	return u
}
