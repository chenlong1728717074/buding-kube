package service

import (
	"sync"
)

var (
	podSrv  *PodService
	podOnce sync.Once
)

type PodService struct {
}

func GetSingletonPodService() *PodService {
	podOnce.Do(func() {
		podSrv = NewPodService()
	})
	return podSrv
}

func NewPodService() *PodService {
	return &PodService{}
}
