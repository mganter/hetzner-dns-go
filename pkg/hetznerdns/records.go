package hetznerdns

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type GetRecordsOpts struct {
	ZoneID  ZoneID
	PerPage int
	Page    int
}

const (
	recordsQueryParamPage    = "page"
	recordsQueryParamPerPage = "per_page"
	recordsQueryParamZoneID  = "zone_id"
)

func (h *dNSHetzner) GetRecords(ctx context.Context, opts GetRecordsOpts) (Records, error) {

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

	requestUrl, err := url.Parse(hetznerDNSBaseURL + "records?" + queryParams.Encode())
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
		bdy, _ := ioutil.ReadAll(response.Body)
		return Records{}, fmt.Errorf("%w: %s %s", ErrUnknown, response.Status, bdy)
	}

	var recordsResponse recordsResponse
	err = json.NewDecoder(response.Body).Decode(&recordsResponse)
	if err != nil {
		return Records{}, fmt.Errorf("%w: %s", ErrUnknown, err.Error())
	}

	return recordsResponse.Records, nil
}
