package health

import (
	"fmt"
	"net/http"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Healthy")
	})

	return mux
}
