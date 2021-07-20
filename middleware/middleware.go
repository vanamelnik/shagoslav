package middleware

import (
	"log"
	"net/http"
)

// RequireMeetingToken tries to find in url queries 'token' field and if not found
// redirects to /group page
func RequireMeetingToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.URL.Query()["token"]
		if !ok {
			log.Println("handler: bad request - token field in query required. Redirecting to /group")
			http.Redirect(w, r, "/group", http.StatusFound)
			return
		}
		next(w, r)
	}
}
