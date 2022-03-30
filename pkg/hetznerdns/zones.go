package hetznerdns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const hetznerDNSBaseURL = "https://dns.hetzner.com/api/v1/"

type ZoneID string

var (
	ErrBadRequest           = errors.New("400 Bad Request")
	ErrUnauthorized         = errors.New("401 Unauthorized")
	ErrForbidden            = errors.New("403 Forbidden")
	ErrNotFound             = errors.New("404 Not Found")
	ErrNotAcceptable        = errors.New("406 Not Acceptable")
	ErrUnprocessableEntity  = errors.New("422 Not Acceptable")
	ErrCouldNotBuildRequest = errors.New("request could not be built")
	ErrUnknown              = errors.New("unknown error")
)

type dNSHetzner struct {
	authToken  string
	httpClient *http.Client
}

func NewDNSClient(authToken string) DNS {
	return &dNSHetzner{
		authToken: authToken,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func NewDNSClientWithHttpClient(authToken string, client *http.Client) DNS {
	return &dNSHetzner{
		authToken:  authToken,
		httpClient: client,
	}
}

func (h *dNSHetzner) CreateZone(ctx context.Context, name string, defaultTTL uint64) (Zone, error) {
	body, err := json.Marshal(struct {
		Name string `json:"name"`
		TTL  uint64 `json:"ttl"`
		TLD  string `json:"tld"`
	}{
		TLD:  "dev",
		Name: name,
		TTL:  defaultTTL,
	})
	if err != nil {
		return Zone{}, fmt.Errorf("%w: %s", ErrCouldNotBuildRequest, err.Error())
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, hetznerDNSBaseURL+"zones", bytes.NewReader(body))
	if err != nil {
		return Zone{}, fmt.Errorf("%w: %s", ErrCouldNotBuildRequest, err.Error())
	}
	h.addHeaderToRequest(request)

	response, err := h.httpClient.Do(request)
	if err != nil {
		return Zone{}, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusCreated:
		// all good
		break
	case http.StatusUnauthorized:
		return Zone{}, ErrUnauthorized
	case http.StatusNotAcceptable:
		panic(fmt.Errorf("%w: maybe content headers were not set properly", ErrNotAcceptable))
	case http.StatusUnprocessableEntity:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s", ErrUnprocessableEntity, bdy)
	default:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var zone zoneResponse
	err = json.NewDecoder(response.Body).Decode(&zone)
	if err != nil {
		return Zone{}, err
	}

	return zone.Zone, nil
}

func (h *dNSHetzner) GetZones(ctx context.Context) (Zones, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, hetznerDNSBaseURL+"zones", nil)
	if err != nil {
		return Zones{}, fmt.Errorf("%w: %s", ErrCouldNotBuildRequest, err.Error())
	}

	h.addHeaderToRequest(request)

	response, err := h.httpClient.Do(request)
	if err != nil {
		return Zones{}, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		// all good
		break
	case http.StatusBadRequest:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zones{}, fmt.Errorf("%w: %s", ErrBadRequest, bdy)
	case http.StatusUnauthorized:
		return Zones{}, fmt.Errorf("%w: %s", ErrUnauthorized, request.URL)
	case http.StatusNotAcceptable:
		panic(fmt.Errorf("%w: maybe content headers were not set properly", ErrNotAcceptable))
	default:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zones{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var zones zonesResponse
	err = json.NewDecoder(response.Body).Decode(&zones)
	if err != nil {
		return Zones{}, err
	}

	return Zones{
		Zones: zones.Zones,
	}, nil
}

func (h *dNSHetzner) GetZone(ctx context.Context, id ZoneID) (Zone, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, hetznerDNSBaseURL+"zones/"+string(id), nil)
	if err != nil {
		return Zone{}, fmt.Errorf("%w: %s", ErrCouldNotBuildRequest, err.Error())
	}

	h.addHeaderToRequest(request)

	response, err := h.httpClient.Do(request)
	if err != nil {
		return Zone{}, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		// all good
		break
	case http.StatusUnauthorized:
		return Zone{}, ErrUnauthorized
	case http.StatusForbidden:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s", ErrForbidden, bdy)
	case http.StatusNotFound:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s", ErrNotFound, bdy)
	case http.StatusNotAcceptable:
		panic(fmt.Errorf("%w: maybe content headers were not set properly", ErrNotAcceptable))
	default:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var zone zoneResponse
	err = json.NewDecoder(response.Body).Decode(&zone)
	if err != nil {
		return Zone{}, err
	}

	return zone.Zone, nil
}
