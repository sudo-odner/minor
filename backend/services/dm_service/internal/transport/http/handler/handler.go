package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type DMChannelService interface{}

type Handler struct {
	log              *slog.Logger
	dmChannelService DMChannelService
	validate         *validator.Validate
}

func New(log *slog.Logger, channelService DMChannelService) *Handler {
	return &Handler{
		log:              log,
		dmChannelService: channelService,
		validate:         validator.New(),
	}
}

func (h *Handler) CreateDialog(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) DeleteDialog(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) ListUserDMChannels(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
}
