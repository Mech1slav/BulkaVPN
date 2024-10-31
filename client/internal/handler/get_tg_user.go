package handler

import (
	"context"

	"BulkaVPN/client/internal/repository"
)

func (h *Handler) GetUserByTelegramID(ctx context.Context, telegramID int64) (*repository.User, error) {
	user, err := h.clientRepo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
