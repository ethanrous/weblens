package routers

import (
	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	v1 "github.com/ethanrous/weblens/routers/api/v1"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/services/context"
)

func startup(ctx context.BasicContext) error {
	r := router.NewRouter()

	cnf := config.GetConfig()

	mongo, err := db.ConnectToMongo(ctx, cnf.MongoDBUri, cnf.MongoDBName)
	if err != nil {
		return err
	}

	r.Inject(router.Injection{
		DB:  mongo,
		Log: ctx.Log,
	})

	r.Mount("/api/v1/", v1.Routes())

	return nil
}
