package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
)

func setupProxy(svr *Service, proxyEndpoints map[string]url.URL) {
	for endpoint, url := range proxyEndpoints {
		svr.Router.PathPrefix("/" + endpoint).HandlerFunc(ProxyRequestHandler(&url, endpoint, svr.GetLogger()))
	}
}

func ProxyRequestHandler(target *url.URL, endpoint string, logger *zerolog.Logger) func(http.ResponseWriter, *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/"+endpoint)
		logger.Info().Msgf("Proxying request to %s%s", target.String(), req.URL.Path)
	}

	return proxy.ServeHTTP
}
