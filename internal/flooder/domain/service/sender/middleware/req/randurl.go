package reqmiddleware

import (
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type RandUrlMiddleware struct {
	URLs   []string
	logger logservice.Logger
}

func NewRandUrlMiddleware(URLs []string, logger logservice.Logger) *RandUrlMiddleware {
	return &RandUrlMiddleware{URLs: URLs, logger: logger}
}

func (m *RandUrlMiddleware) AddRandUrl(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
	return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		if req != nil {
			rand.Seed(time.Now().UnixNano())

			if len(m.URLs) != 0 {
				var u string
				if len(m.URLs) > 1 {
					u = m.URLs[rand.Intn(len(m.URLs))]
				} else {
					u = m.URLs[0]
				}
				p, err := url.Parse(u)
				if err != nil {
					m.logger.Println("unable to parse given url, " + err.Error())
					return next.Do(req)
				}
				req.URL = p
			}
		}

		return next.Do(req)
	})
}
