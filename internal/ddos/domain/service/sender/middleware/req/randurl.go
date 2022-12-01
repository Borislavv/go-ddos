package reqmiddleware

import (
	ddos "ddos/internal/ddos/app"
	"ddos/internal/ddos/infrastructure/httpclient"
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func RandUrl(next httpclient.RequestModifier) httpclient.RequestModifier {
	return httpclient.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		rand.Seed(time.Now().UnixNano())

		if len(ddos.URLs) != 0 {
			u := ddos.URLs[rand.Intn(len(ddos.URLs)+1)]
			p, err := url.Parse(u)
			if err != nil {
				return nil, errors.New("unable to parse given url, " + err.Error())
			}
			req.URL = p
		}

		return next.Do(req)
	})
}
