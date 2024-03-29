package server

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/Safety-Third/prismriver/internal/app/constants"
	"github.com/Safety-Third/prismriver/internal/app/server/routes/media"
	"github.com/Safety-Third/prismriver/internal/app/server/routes/player"
	"github.com/Safety-Third/prismriver/internal/app/server/routes/queue"
	"github.com/Safety-Third/prismriver/internal/app/server/routes/queue/item"
	"github.com/Safety-Third/prismriver/internal/app/server/ws/routes"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// CreateRouter creates the Router instance used for handling all incoming HTTP requests.
func CreateRouter() {
	wait := time.Duration(15)

	r := mux.NewRouter()
	r.HandleFunc("/media", media.IndexHandler).Methods("GET")
	r.HandleFunc("/media/{type}/{id}", media.UpdateHandler).Methods("PUT")
	r.HandleFunc("/player", player.UpdateHandler).Methods("PUT")
	r.HandleFunc("/queue", queue.IndexHandler).Methods("GET")
	r.HandleFunc("/queue", queue.StoreHandler).Methods("POST")
	r.HandleFunc("/queue", queue.UpdateHandler).Methods("PUT")
	r.HandleFunc("/queue/{id}", item.DeleteHandler).Methods("DELETE")
	r.HandleFunc("/queue/{id}", item.UpdateHandler).Methods("PUT")
	r.HandleFunc("/ws/player", routes.WebsocketPlayerHandler)
	r.HandleFunc("/ws/queue", routes.WebsocketQueueHandler)

	srv := &http.Server{
		Addr: ":8000",
		Handler: handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
			handlers.AllowedOrigins([]string{viper.GetString(constants.ORIGIN)}))(r),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logrus.Fatalf("could not start http server: %v", err)
		}
	}()
	logrus.Info("HTTP server now listening on port 8000.")

	// we call the methods used for retrieving the websocket instances to guarantee that they are instantiated in order
	// to prevent potential deadlocks caused by the player and queue sending updates on their channels that aren't
	// being handled
	routes.GetPlayerHub()
	routes.GetQueueHub()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	logrus.Info("HTTP server gracefully shut down.")
}
