package service

import (
	"sync"
)

var (
	toolSrv  *ToolService
	toolOnce sync.Once
)

type ToolService struct {
}

func NewToolService() *ToolService {
	return &ToolService{}
}

func GetSingletonToolService() *ToolService {
	toolOnce.Do(func() {
		toolSrv = NewToolService()
	})
	return toolSrv
}
