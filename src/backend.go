//go:generate oapi-codegen --exclude-tags=session -generate types,gin-server -o api/api.gen.go -package api api/openapi.yaml

package main

import (
	"OPP/backend/api"
	"OPP/backend/auth"
	"OPP/backend/db"
	"OPP/backend/handlers"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

type opp_handlers struct {
	handlers.UserHandlers
	handlers.CarHandlers
	handlers.TicketHandlers
	handlers.FineHandlers
}

var DEBUG_MODE = os.Getenv("DEBUG_MODE") == "true"

func main() {

	// Initialize the database connection
	if db.Init() != nil {
		panic("Failed to initialize database")
	}

	opp_handlers := &opp_handlers{
		UserHandlers:   *handlers.NewUserHandler(),
		CarHandlers:    *handlers.NewCarHandler(),
		TicketHandlers: *handlers.NewTicketHandler(),
		FineHandlers:   *handlers.NewFineHandler(),
	}

	auth := auth.NewAuth()
	if auth == nil {
		panic("Failed to initialize auth")
	}

	// options

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Load OpenAPI spec for validation
	// oapi-codegen do not handle validation from the spec
	// nor authentication
	spec, err := util.LoadSwagger("api/openapi.yaml")
	if err != nil {
		log.Panicf("Failed to load OpenAPI spec: %v", err)
	}

	silenceServersWarning := false
	if DEBUG_MODE {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
		silenceServersWarning = true
	}

	// Set up the authentication function
	validatorOptions := &ginmiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
				return auth.AuthenticationFunc(ctx, input)
			},
		},
		SilenceServersWarning: silenceServersWarning,
	}
	validator := ginmiddleware.OapiRequestValidatorWithOptions(spec, validatorOptions)
	if err != nil {
		log.Panicf("Failed to create validator: %v", err)
	}
	r.Use(validator)
	r.SetTrustedProxies(nil)

	options := api.GinServerOptions{
		BaseURL:      "/api/v1",
		Middlewares:  nil,
		ErrorHandler: nil,
	}
	var opp_h = opp_handlers
	api.RegisterHandlersWithOptions(r, opp_h, options)

	fmt.Println("OPP Backend starting on :8080")
	log.Fatal(r.Run(":8080"))
}
