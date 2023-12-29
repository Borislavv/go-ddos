package reqmiddleware

import (
	"ddos/internal/ddos/infrastructure/httpclient/middleware"
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var URLs = []string{}

func RandUrl(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
	return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		rand.Seed(time.Now().UnixNano())

		if len(URLs) != 0 {
			u := URLs[rand.Intn(len(URLs)+1)]
			p, err := url.Parse(u)
			if err != nil {
				return nil, errors.New("unable to parse given url, " + err.Error())
			}
			req.URL = p
		}

		return next.Do(req)
	})
}
