package service

import "sync"

var (
	kubeSrv     *KubeSrvService
	kubeSrvOnce sync.Once
)

type KubeSrvService struct {
}

func NewKubeSrvService() *KubeSrvService {
	return &KubeSrvService{}
}

func GetSingletonKubeSrvService() *KubeSrvService {
	kubeSrvOnce.Do(func() {
		kubeSrv = NewKubeSrvService()
	})
	return kubeSrv
}
