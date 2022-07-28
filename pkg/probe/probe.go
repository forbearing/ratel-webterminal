package probe

import "net/http"

// HandleHealthyProbe handle api "/-/healthy"
// it alwasy response wth 200 status code and message "ok" if the ratel-webterminal is running.
func HandleHealthyProbe(w http.ResponseWriter, r *http.Request) {
	// you should alwasy call w.WriteHeader before anything else it will output
	// some unexpected message like "http: superfluous response.WriteHeader call from github.com...."
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// HandleReadyProbe handle api "/-/ready"
// it alwasy response wth 200 status code and message "ok" if the ratel-webterminal is running.
func HandleReadyProbe(w http.ResponseWriter, r *http.Request) {
	// you should alwasy call w.WriteHeader before anything else it will output
	// some unexpected message like "http: superfluous response.WriteHeader call from github.com...."
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
