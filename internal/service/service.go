package service

import "github.com/chestorix/gophermart/internal/interfaces"

type Service struct {
	repo interfaces.Repository
}

func NewService(repo interfaces.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Test() string {
	return s.repo.Test()
}
