package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	port   string
	server http.Server
	router *mux.Router
	wg     sync.WaitGroup
}

func NewAdminServer(port int) *Server {
	router := mux.NewRouter().StrictSlash(false)
	return &Server{
		router: router,
		port: fmt.Sprintf(":%d", port),
	}
}

func NewServer(port int) *Server {
	router := mux.NewRouter().StrictSlash(true)
	return &Server{
		router: router,
		port:   fmt.Sprintf(":%d", port),
	}
}

func (c *Server) AddRoute(path string, handler http.HandlerFunc, method string, mwf ...mux.MiddlewareFunc) {
	subRouter := c.router.PathPrefix(path).Subrouter()
	subRouter.Use(mwf...)
	subRouter.HandleFunc("", handler).Methods(method)
	log.Printf("Added route: [%v] [%v]", path, method)
}

func (c *Server) MustStart() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the HTML Server
	c.server = http.Server{
		Addr:           fmt.Sprintf("0.0.0.0%s", c.port),
		Handler:        c.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   0 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}

	c.wg.Add(1)

	// Start the listener
	go func() {
		log.Printf("API server started at %v on http://%s", time.Now().Format(time.Stamp), c.server.Addr)
		if err := c.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("API server failed to start with error: %v\n", err)
		}
		log.Println("API server stopped")
		c.wg.Done()
	}()
}

// make another api server for admins here 

// Stop stops the API Server
func (c *Server) Stop() error {
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Println("API server: stopping")


	if err := c.server.Shutdown(ctx); err != nil {
		if err := c.server.Close(); err != nil {
			log.Printf("API server: stopped with error %v", err)
			return err
		}
		return err
	}

	c.wg.Wait()
	return nil
}