package dashboard

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/dashboard/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// is logged in middleware

}
