package service

import "sync"

var (
	namespaceSrv  *NamespaceService
	namespaceOnce sync.Once
)

type NamespaceService struct {
}

func GetSingletonNamespaceService() *NamespaceService {
	namespaceOnce.Do(func() {
		namespaceSrv = NewNamespaceService()
	})
	return namespaceSrv
}

func NewNamespaceService() *NamespaceService {
	return &NamespaceService{}
}
