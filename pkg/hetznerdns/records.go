package hetznerdns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type RecordType string

const (
	A     = RecordType("A")
	AAAA  = RecordType("AAAA")
	NS    = RecordType("NS")
	MX    = RecordType("MX")
	CNAME = RecordType("CNAME")
	RP    = RecordType("RP")
	TXT   = RecordType("TXT")
	SOA   = RecordType("SOA")
	HINFO = RecordType("HINFO")
	SRV   = RecordType("SRV")
	DANE  = RecordType("DANE")
	TLSA  = RecordType("TLSA")
	DS    = RecordType("DS")
	CAA   = RecordType("CAA")
)

type GetRecordsOpts struct {
	ZoneID  ZoneID
	PerPage int
	Page    int
}

func (opts GetRecordsOpts) asUrlValues() url.Values {
	queryParams := url.Values{}
	if opts.Page != 0 {
		queryParams.Add(recordsQueryParamPage, strconv.Itoa(opts.Page))
	}
	if opts.PerPage != 0 {
		queryParams.Add(recordsQueryParamPerPage, strconv.Itoa(opts.PerPage))
	}
	if opts.ZoneID != "" {
		queryParams.Add(recordsQueryParamZoneID, string(opts.ZoneID))
	}
	return queryParams
}

type CreateRecordOpts struct {
	Name   string     `json:"name"`
	TTL    uint64     `json:"ttl,omitempty"`
	Type   RecordType `json:"type"`
	Value  string     `json:"value"`
	ZoneID string     `json:"zone_id"`
}

const (
	recordsQueryParamPage    = "page"
	recordsQueryParamPerPage = "per_page"
	recordsQueryParamZoneID  = "zone_id"
)

func (h *dNSHetzner) GetRecords(ctx context.Context, opts GetRecordsOpts) (Records, error) {
	requestUrl, err := url.Parse(hetznerDNSBaseURL + "records?" + opts.asUrlValues().Encode())
	if err != nil {
		panic(err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl.String(), nil)
	if err != nil {
		return Records{}, fmt.Errorf("%w: %s", ErrCouldNotBuildRequest, err.Error())
	}

	h.addHeaderToRequest(request)

	response, err := h.httpClient.Do(request)
	if err != nil {
		return Records{}, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		// all good
		break
	case http.StatusUnauthorized:
		return Records{}, fmt.Errorf("%w: %s", ErrUnauthorized, request.URL)
	case http.StatusNotAcceptable:
		panic(fmt.Errorf("%w: maybe content headers were not set properly", ErrNotAcceptable))
	default:
		bdy, _ := io.ReadAll(response.Body)
		return Records{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var recordsResponse recordsResponse
	err = json.NewDecoder(response.Body).Decode(&recordsResponse)
	if err != nil {
		return Records{}, fmt.Errorf("%w: %s", ErrUnknown, err.Error())
	}

	return recordsResponse.Records, nil
}

func (h *dNSHetzner) CreateRecord(ctx context.Context, opts CreateRecordOpts) (Record, error) {
	requestUrl, err := url.Parse(hetznerDNSBaseURL + "records")
	if err != nil {
		panic(err)
	}

	requestBody, err := json.Marshal(opts)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestUrl.String(), bytes.NewReader(requestBody))
	if err != nil {
		return Record{}, fmt.Errorf("%w: %s", ErrCouldNotBuildRequest, err.Error())
	}

	h.addHeaderToRequest(request)

	response, err := h.httpClient.Do(request)
	if err != nil {
		return Record{}, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		// all good
		break
	case http.StatusUnauthorized:
		return Record{}, fmt.Errorf("%w: %s", ErrUnauthorized, request.URL)
	case http.StatusForbidden:
		return Record{}, fmt.Errorf("%w: %s", ErrForbidden, request.URL)
	case http.StatusNotAcceptable:
		panic(fmt.Errorf("%w: maybe content headers were not set properly", ErrNotAcceptable))
	case http.StatusUnprocessableEntity:
		return Record{}, fmt.Errorf("%w: %s", ErrUnprocessableEntity, request.URL)
	default:
		bdy, _ := io.ReadAll(response.Body)
		return Record{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var createRecordResponse createRecordResponse
	err = json.NewDecoder(response.Body).Decode(&createRecordResponse)
	if err != nil {
		return Record{}, fmt.Errorf("%w: %s", ErrUnknown, err.Error())
	}

	return createRecordResponse.Record, nil
}
