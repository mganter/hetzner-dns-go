package hetznerdns

import (
	"net/http"
)

func (h *dNSHetzner) addHeaderToRequest(request *http.Request) {
	request.Header.Add("Auth-API-Token", h.authToken)
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
}
