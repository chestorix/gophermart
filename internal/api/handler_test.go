package api

import (
	"context"
	"github.com/chestorix/gophermart/internal/config"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"testing"
)

func TestHandler_GetTest(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.GetTest(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_GetUserBalance(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.GetUserBalance(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_GetUserOrders(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.GetUserOrders(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_GetUserWithdrawals(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.GetUserWithdrawals(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.Login(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_Register(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.Register(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_UploadOrder(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.UploadOrder(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_Withdraw(t *testing.T) {
	type fields struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
				dbURL:   tt.fields.dbURL,
			}
			h.Withdraw(tt.args.w, tt.args.r)
		})
	}
}

func TestNewHandler(t *testing.T) {
	type args struct {
		service interfaces.Service
		logger  *logrus.Logger
		dbURL   string
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.service, tt.args.logger, tt.args.dbURL); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRouter(t *testing.T) {
	type args struct {
		logger *logrus.Logger
	}
	tests := []struct {
		name string
		args args
		want *Router
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRouter(tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		cfg     *config.ServerConfig
		service interfaces.Service
		logger  *logrus.Logger
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.cfg, tt.args.service, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_SetupRoutes(t *testing.T) {
	type fields struct {
		Router chi.Router
		logger *logrus.Logger
	}
	type args struct {
		handler *Handler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Router{
				Router: tt.fields.Router,
				logger: tt.fields.logger,
			}
			r.SetupRoutes(tt.args.handler)
		})
	}
}

func TestServer_Shutdown(t *testing.T) {
	type fields struct {
		cfg     *config.ServerConfig
		router  *Router
		service interfaces.Service
		server  *http.Server
		logger  *logrus.Logger
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:     tt.fields.cfg,
				router:  tt.fields.router,
				service: tt.fields.service,
				server:  tt.fields.server,
				logger:  tt.fields.logger,
			}
			if err := s.Shutdown(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Start(t *testing.T) {
	type fields struct {
		cfg     *config.ServerConfig
		router  *Router
		service interfaces.Service
		server  *http.Server
		logger  *logrus.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:     tt.fields.cfg,
				router:  tt.fields.router,
				service: tt.fields.service,
				server:  tt.fields.server,
				logger:  tt.fields.logger,
			}
			if err := s.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
