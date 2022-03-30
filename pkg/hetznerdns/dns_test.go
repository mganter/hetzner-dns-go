package hetznerdns_test

import (
	"bytes"
	"context"
	"github.com/ganterm/hetzner-dns-go/pkg/hetznerdns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"testing"
)

func TestDNSClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Books Suite")
}

var _ = Describe("GetRecords", func() {
	Context("With no optional parameter", func() {
		var request *http.Request
		var records hetznerdns.Records
		var err error
		BeforeEach(func() {
			client := NewTestClient(func(req *http.Request) *http.Response {
				request = req

				return &http.Response{
					Status:     "OK",
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewReader([]byte(`
{
  "records": [
    {
      "type": "A",
      "zone_id": "my-id",
      "name": "string"
    }
  ]
}`))),
				}
			})
			dnsClient := hetznerdns.NewDNSClientWithHttpClient("my-token", client)
			records, err = dnsClient.GetRecords(context.TODO(), hetznerdns.GetRecordsOpts{})
		})

		It("to have an empty query parameter", func() {
			Expect(request.URL.Query()).To(BeEmpty())
		})

		It("to have a recordset with one entry", func() {
			Expect(records).To(HaveLen(1))
			Expect(records[0].Type).To(Equal("A"))
			Expect(records[0].ZoneId).To(Equal("my-id"))
			Expect(records[0].Name).To(Equal("string"))
		})

		It("to have not an error", func() {
			Expect(err).To(BeNil())
		})
	})

})

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}
