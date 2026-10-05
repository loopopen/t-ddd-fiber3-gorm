package http

import (
	"log/slog"
	"time"

	"github.com/gofiber/contrib/v3/swagger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/loopopen/t-ddd-fiber3-gorm/cmd/api/config"
	"github.com/loopopen/t-ddd-fiber3-gorm/docs"
	v1 "github.com/loopopen/t-ddd-fiber3-gorm/internal/adapters/http/handlers/v1"
	timeout2 "github.com/loopopen/t-ddd-fiber3-gorm/internal/adapters/http/timeout"
	"github.com/loopopen/t-ddd-fiber3-gorm/internal/applic/service"

	"github.com/gofiber/fiber/v3/middleware/timeout"
	"github.com/google/wire"
	"github.com/loopopen/gap"
)

var ProviderSet = wire.NewSet(NewApp)

func NewApp(
	e *config.Env,
	c *config.Config,
	userSvc *service.UserSvc,
	pub gap.EventPublisher,
	logger *slog.Logger,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: HandlerError(logger),
		BodyLimit:    4 * 1024 * 1024,
	})
	if !e.IsProd() && c.Timeout > 0 {
		app.Use(timeout.New(func(c fiber.Ctx) error {
			return c.Next()
		}, timeout.Config{Timeout: time.Duration(c.Timeout) * time.Millisecond}))
	}
	app.Use(recover.New(recover.Config{
		EnableStackTrace: !e.IsProd(),
	}))
	corsConf := cors.ConfigDefault
	if len(c.CORS.AllowOrigins) > 0 {
		corsConf.AllowOrigins = c.CORS.AllowOrigins
	}
	if len(c.CORS.AllowHeaders) > 0 {
		corsConf.AllowHeaders = c.CORS.AllowHeaders
	}
	app.Use(cors.New(corsConf))

	docs.SwaggerInfo.Host = c.Swagger.Host
	docs.SwaggerInfo.BasePath = c.Swagger.BasePath
	docs.SwaggerInfo.Title += "-" + e.Name
	if e.CommitSHA != "" {
		docs.SwaggerInfo.Version += "-" + e.CommitSHA
	}
	app.Use(swagger.New(swagger.Config{
		FileContent: []byte(docs.SwaggerInfo.ReadDoc()),
		Path:        "swagger",
		Title:       docs.SwaggerInfo.Title,
	}))

	if pub != nil {
		app.All("/dashboard/*", adaptor.HTTPHandler(gap.NewDashboardHandler(pub)))
	}

	apiv1 := app.Group("/api/v1")

	users := apiv1.Group("/users")
	users.Get("", v1.QueryUsers(userSvc))
	users.Get("unsafe-timeout", timeout2.UnsafeTimeout(v1.QueryUsersWithUnsafeTimeout(userSvc), 1*time.Second))

	if pub != nil {
		gap.Subscribe(
			gap.From(pub),
			gap.Inject(userSvc),
		)
	}
	return app
}
