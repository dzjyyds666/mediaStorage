package server

import "github.com/dzjyyds666/vortex/v2"

type StorageHandler struct {
}

func NewStorageHandler() *StorageHandler { return &StorageHandler{} }

func (s *StorageHandler) HandleRouterTest(ctx *vortex.Context) error {
	return nil
}
