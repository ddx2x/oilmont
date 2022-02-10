package gateway

import (
	"net/http"
	"strings"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/micro/gateway"
	"github.com/ddx2x/oilmont/pkg/utils/token"
)

type InterceptMiddleware interface {
	Intercept() gateway.Intercept
}

// LoginHandle processing the login request, if not, then the middleware processing later
func LoginHandle(w http.ResponseWriter, r *http.Request) gateway.InterceptType {
	if strings.HasPrefix(r.URL.String(), LoginURL) {
		return gateway.SelfHandle
	}
	if strings.HasPrefix(r.URL.String(), FeiShuLoginURL) {
		return gateway.SelfHandle
	}
	if strings.HasPrefix(r.URL.String(), SHELL) {
		return gateway.Redirect
	}
	if strings.HasPrefix(r.URL.String(), WatchURL) {
		authorizedHeader := r.Header.Get(common.AuthorizationHeader)
		if authorizedHeader == "" {
			return gateway.NotAuthorized
		}
		cc, err := token.Decode(authorizedHeader)
		if err != nil {
			return gateway.NotAuthorized
		}
		//check token
		r.Header.Set(common.HttpRequestUserHeaderKey, cc.UserName)
		// if watch use NativeEventSource get url
		return gateway.SelfHandle
	}

	return gateway.Next
}

func JWTToken(w http.ResponseWriter, r *http.Request) gateway.InterceptType {
	authorizedHeader := r.Header.Get(common.AuthorizationHeader)
	if authorizedHeader == "" {
		return gateway.NotAuthorized
	}

	cc, err := token.Decode(authorizedHeader)
	if err != nil {
		return gateway.NotAuthorized
	}
	//check token
	r.Header.Set(common.HttpRequestUserHeaderBELONGTENANT, cc.Issuer)
	r.Header.Set(common.HttpRequestUserHeaderKey, cc.UserName)
	return gateway.Next
}

func DefaultRedirect(w http.ResponseWriter, r *http.Request) gateway.InterceptType {
	return gateway.Redirect
}
