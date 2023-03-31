package main

import (
	"net/http"
	"os"
	"time"

	"github.com/energietransitie/twomes-backoffice-api/handlers"
	"github.com/energietransitie/twomes-backoffice-api/pkg/repositories"
	"github.com/energietransitie/twomes-backoffice-api/pkg/services"
	"github.com/energietransitie/twomes-backoffice-api/pkg/twomes"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// Configuration holds all the configuration for the server.
type Configuration struct {
	DatabaseDSN string
}

func getConfiguration() Configuration {
	dsn, ok := os.LookupEnv("TWOMES_DSN")
	if !ok {
		logrus.Fatal("TWOMES_DSN was not set")
	}

	return Configuration{
		DatabaseDSN: dsn,
	}
}

func main() {
	config := getConfiguration()

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	db, err := repositories.NewDatabaseConnectionAndMigrate(config.DatabaseDSN)
	if err != nil {
		logrus.Fatal(err)
	}

	authService, err := services.NewAuthorizationServiceFromFile("./data/key.pem")
	if err != nil {
		logrus.Fatal(err)
	}
	authHandler := handlers.NewAuthorizationHandler(authService)

	// Print an admin authorization token.
	{
		adminAuthToken, err := authService.CreateTokenFromAuthorization(twomes.Authorization{
			Kind: twomes.AdminToken,
			ID:   0,
		})
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Info("admin authorization token: ", adminAuthToken)
	}

	adminAuth := authHandler.Middleware(twomes.AdminToken)
	accountActivationAuth := authHandler.Middleware(twomes.AccountActivationToken)
	accountAuth := authHandler.Middleware(twomes.AccountToken)
	deviceAuth := authHandler.Middleware(twomes.DeviceToken)

	appRepository := repositories.NewAppRepository(db)
	campaignRepository := repositories.NewCampaignRepository(db)
	buildingRepository := repositories.NewBuildingRepository(db)
	accountRepository := repositories.NewAccountRepository(db)
	propertyRepository := repositories.NewPropertyRepository(db)
	deviceTypeRepository := repositories.NewDeviceTypeRepository(db)
	deviceRepository := repositories.NewDeviceRepository(db)
	uploadRepository := repositories.NewUploadRepository(db)

	appService := services.NewAppService(appRepository)
	campaignService := services.NewCampaignService(campaignRepository, appService)
	buildingService := services.NewBuildingService(buildingRepository)
	accountService := services.NewAccountService(accountRepository, authService, appService, campaignService, buildingService)
	propertyService := services.NewPropertyService(propertyRepository)
	deviceTypeService := services.NewDeviceTypeService(deviceTypeRepository, propertyService)
	deviceService := services.NewDeviceService(deviceRepository, authService, deviceTypeService, buildingService)
	uploadService := services.NewUploadService(uploadRepository, propertyService)

	appHandler := handlers.NewAppHandler(appService)
	campaignHandler := handlers.NewCampaignHandler(campaignService)
	accountHandler := handlers.NewAccountHandler(accountService)
	propertyHandler := handlers.NewPropertyHandler(propertyService)
	deviceTypeHandler := handlers.NewDeviceTypeHandler(deviceTypeService)
	deviceHandler := handlers.NewDeviceHandler(deviceService)
	uploadHandler := handlers.NewUploadHandler(uploadService)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logrus.StandardLogger()}))
	r.Use(middleware.Timeout(time.Second * 30))

	r.Method("POST", "/app", adminAuth(appHandler.Create)) // POST on /app.

	r.Method("POST", "/campaign", adminAuth(campaignHandler.Create)) // POST on /campaign.

	r.Route("/account", func(r chi.Router) {
		r.Method("POST", "/", adminAuth(accountHandler.Create))                       // POST on /account.
		r.Method("POST", "/activate", accountActivationAuth(accountHandler.Activate)) // POST on /account/activate.

	})

	r.Method("POST", "/property", adminAuth(propertyHandler.Create)) // POST on /property.

	r.Method("POST", "/device_type", adminAuth(deviceTypeHandler.Create)) // POST on /device_type.

	r.Route("/device", func(r chi.Router) {
		r.Method("POST", "/", accountAuth(deviceHandler.Create))                      // POST on /device.
		r.Method("POST", "/activate", handlers.Handler(deviceHandler.Activate))       // POST on /device/activate.
		r.Method("GET", "/{device_name}", accountAuth(deviceHandler.GetDeviceByName)) // GET on /device/{device_name}.
	})

	r.Method("POST", "/upload", deviceAuth(uploadHandler.Create)) // POST on /upload.

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		logrus.Fatal(err)
	}
}