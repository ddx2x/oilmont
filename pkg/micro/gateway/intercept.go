package gateway

import (
	"github.com/micro/micro/v2/plugin"
	"net/http"
)

func ServerIntercept(self http.Handler, intercepts ...Intercept) plugin.Handler {
	return func(redirect http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state := Redirect
			for _, intercept := range intercepts {
				state = intercept(w, r)
				if state != Next {
					break
				}
			}

			switch state {
			case SelfHandle:
				self.ServeHTTP(w, r)
			case Interrupt:
				w.WriteHeader(http.StatusNotFound)
				return
			case NotAuthorized:
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("access denied"))
				return
			default:
				redirect.ServeHTTP(w, r)
			}
		})
	}
}
