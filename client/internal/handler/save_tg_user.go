package handler

import (
	"context"

	"BulkaVPN/client/internal/repository"
)

func (h *Handler) SaveUser(ctx context.Context, user *repository.User) error {
	err := h.clientRepo.SaveUser(ctx, user)
	if err != nil {
		return err
	}
	return nil
}
