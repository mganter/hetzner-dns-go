package hetznerdns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const hetznerDNSBaseURL = "https://dns.hetzner.com/api/v1/"

type ZoneID string

var (
	ErrBadRequest                           = errors.New("400 Bad Request")
	ErrPaginationSelectorsMutuallyExclusive = errors.New("400 Pagination selectors are mutually exclusive")
	ErrUnauthorized                         = errors.New("401 Unauthorized")
	ErrForbidden                            = errors.New("403 Forbidden")
	ErrNotFound                             = errors.New("404 Not Found")
	ErrNotAcceptable                        = errors.New("406 Not Acceptable")
	ErrUnprocessableEntity                  = errors.New("422 Not Acceptable")
	ErrCouldNotBuildRequest                 = errors.New("request could not be built")
	ErrUnknown                              = errors.New("unknown error")
)

type ListZoneOpts struct {
	// if Page is 0, it will be considered as the APIs default value (100)
	Page uint64
	// if PerPage is 0, it will be considered as the APIs default value (100). Must not be bigger than 100.
	PerPage uint64
	// returns the zone with the name or returns 404
	Name string
	// Partial name of a zone
	SearchName string
}

func (opts ListZoneOpts) asQueryParams() string {
	var queryParams []string
	if opts.Page != 0 {
		queryParams = append(queryParams, fmt.Sprintf("page=%v", opts.Page))
	}
	if opts.PerPage != 0 {
		queryParams = append(queryParams, fmt.Sprintf("per_page=%v", opts.PerPage))
	}
	if opts.Name != "" {
		queryParams = append(queryParams, fmt.Sprintf("name=%v", opts.Name))
	}
	if opts.SearchName != "" {
		queryParams = append(queryParams, fmt.Sprintf("search_name=%v", opts.SearchName))
	}
	return strings.Join(queryParams, "&")
}

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
		bdy, _ := io.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s", ErrUnprocessableEntity, bdy)
	default:
		bdy, _ := io.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var zone zoneResponse
	err = json.NewDecoder(response.Body).Decode(&zone)
	if err != nil {
		return Zone{}, err
	}

	return zone.Zone, nil
}

func (h *dNSHetzner) GetZones(ctx context.Context, opts ListZoneOpts) (Zones, Meta, error) {
	url := fmt.Sprintf("%v/zones?%v", hetznerDNSBaseURL, opts.asQueryParams())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Zones{}, Meta{}, fmt.Errorf("%w: %s", ErrCouldNotBuildRequest, err.Error())
	}

	h.addHeaderToRequest(request)

	response, err := h.httpClient.Do(request)
	if err != nil {
		return Zones{}, Meta{}, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		// all good
		break
	case http.StatusBadRequest:
		bdy, _ := io.ReadAll(response.Body)
		return Zones{}, Meta{}, fmt.Errorf("%w: %s", ErrBadRequest, bdy)
	case http.StatusUnauthorized:
		return Zones{}, Meta{}, fmt.Errorf("%w: %s", ErrUnauthorized, request.URL)
	case http.StatusNotAcceptable:
		panic(fmt.Errorf("%w: maybe content headers were not set properly", ErrNotAcceptable))
	default:
		bdy, _ := ioutil.ReadAll(response.Body)
		return Zones{}, Meta{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var zones zonesResponse
	err = json.NewDecoder(response.Body).Decode(&zones)
	if err != nil {
		return Zones{}, Meta{}, err
	}

	return Zones{
		Zones: zones.Zones,
	}, zones.Meta, nil
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
		body, _ := io.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s", ErrForbidden, body)
	case http.StatusBadRequest:

	case http.StatusNotFound:
		body, _ := io.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s", ErrNotFound, body)
	case http.StatusNotAcceptable:
		panic(fmt.Errorf("%w: maybe content headers were not set properly", ErrNotAcceptable))
	default:
		body, _ := io.ReadAll(response.Body)
		return Zone{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, body)
	}

	var zone zoneResponse
	err = json.NewDecoder(response.Body).Decode(&zone)
	if err != nil {
		return Zone{}, err
	}

	return zone.Zone, nil
}
