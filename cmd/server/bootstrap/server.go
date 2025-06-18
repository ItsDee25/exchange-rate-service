package bootstrap

import (
	"context"
	"net/http"
	"time"

	"github.com/ItsDee25/exchange-rate-service/cmd/server/bootstrap/builders"
	infra "github.com/ItsDee25/exchange-rate-service/infra/ratefetcher"
	constants "github.com/ItsDee25/exchange-rate-service/internal/constants/currency"
	repository "github.com/ItsDee25/exchange-rate-service/internal/repository/currency"
	"github.com/ItsDee25/exchange-rate-service/internal/router"
	usecase "github.com/ItsDee25/exchange-rate-service/internal/usecase/currency"
	jobs "github.com/ItsDee25/exchange-rate-service/jobs/currency"
	"github.com/ItsDee25/exchange-rate-service/mocks"
	pkg "github.com/ItsDee25/exchange-rate-service/pkg/awsclient"
	"github.com/gin-gonic/gin"
)

func InitServer() {

	r := gin.Default()

	ctx := context.Background()

	dynamoClient, err := pkg.NewDynamoClient(ctx)
	if err != nil {
		panic("Failed to initialize DynamoDB client: " + err.Error())
	}

	// build env
	env := builders.NewEnv().
		WithDynamoClient(dynamoClient).
		WithHTTPClient(&http.Client{Timeout: 10 * time.Second})

	// build repositories
	cache := repository.NewRateCache()
	repositories := builders.NewRepositories().
		WithCurrencyCache(cache).
		WithCurrencyRepository(repository.NewDynamoRepository(env.DynamoClient, mocks.NewMockRateFetcher(), cache)).
		WithDynamoLocker(infra.NewDynamoLocker(env.DynamoClient))

	// build usecases

	usecases := builders.NewUsecases().
		WithCurrencyUsecase(usecase.NewCurrencyUsecase(repositories.CurrencyDynamoRepository))

	router.RegisterRoutes(r, usecases)

	// start cron jobs

	refresher := jobs.NewRateRefresher(
		repositories.CurrencyDynamoRepository,
		mocks.NewMockRateFetcher(),
		repositories.DynamoLocker,
		constants.SupportedCurrencyPairs,
	)
	refresher.Start()

	cacheCleaner := jobs.NewCacheCleaner(repositories.CurrecyCache)

	cacheCleaner.Start()

	if err := r.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
