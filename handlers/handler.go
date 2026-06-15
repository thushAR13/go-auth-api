package handlers

type Handler struct {
	store Storer
}

func NewHandler(store Storer) *Handler {
	return &Handler{store: store}
}
