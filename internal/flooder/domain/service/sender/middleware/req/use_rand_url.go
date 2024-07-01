package reqmiddleware

import (
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type UseRandUrlMiddleware struct {
	urls   []string
	logger logservice.Logger
}

func NewUseRandUrlMiddleware(URLs []string, logger logservice.Logger) *UseRandUrlMiddleware {
	return &UseRandUrlMiddleware{urls: URLs, logger: logger}
}

func (m *UseRandUrlMiddleware) UseRandUrl(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
	return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		if req != nil {
			random := rand.New(rand.NewSource(time.Now().UnixNano()))

			if len(m.urls) != 0 {
				var u *url.URL
				var err error
				if len(m.urls) > 1 {
					u, err = url.Parse(m.urls[random.Intn(len(m.urls))])
				} else {
					u, err = url.Parse(m.urls[0])
				}
				if err != nil {
					m.logger.Println("unable to parse given url, " + err.Error())
					return next.Do(req)
				}
				req.URL = u
			}
		}

		return next.Do(req)
	})
}
