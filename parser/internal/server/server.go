package server

import (
	"context"
	"log"
	"net/http"
	"parser/configs"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	config     *configs.ServerConfig
}

// Конструктор для сервера
func NewServer(ctx context.Context, config *configs.ServerConfig) (*Server, error) {
	// создаём экземпляр роутера
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}

	// Добавляем middleware для проброса контекста
	router.Use(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "request_id", c.GetHeader("X-Request-ID"))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	return &Server{
		router: router,
		config: config,
	}, nil
}

// Метод для маршрутизации сервера
func (s *Server) SetUpRoutes() {
	s.router.POST("/")
}

// Метод для запуска сервера
func (s *Server) Run() error {
	s.SetUpRoutes()

	s.httpServer = &http.Server{
		Addr:    s.config.Addr(),
		Handler: s.router,
	}
	log.Println("Server is running on port 8080")
	return s.httpServer.ListenAndServe()
}

// Метод для graceful shutdown
func (s *Server) Shutdown(ctx context.Context) error {

	// Останавливаем HTTP сервер
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server shutdown completed")
	return nil
}
