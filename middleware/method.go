// method middleware
package middleware

import (
	"net/http"
)

// method checking middleware
func MethodChecker(allowedMethods []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check if the request method is allowed in the methods
		methodAllowed := false
		for _, method := range allowedMethods {
			if r.Method == method {
				methodAllowed = true
				break
			}
		}
		if !methodAllowed {
			http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		}
		next.ServeHTTP(w, r)
	})
}
