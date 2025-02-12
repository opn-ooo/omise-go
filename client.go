package omise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"mime/multipart"

	"github.com/gorilla/schema"
	"github.com/omise/omise-go/internal"
)

var encoder = schema.NewEncoder()

// Client helps you configure and perform HTTP operations against Omise's REST API. It
// should be used with operation structures from the operations subpackage.
type Client struct {
	*http.Client
	debug bool
	pkey  string
	skey  string
	ckey  string

	// Overrides
	Endpoints map[internal.Endpoint]string

	// configuration
	APIVersion string
	GoVersion  string
}

// NewClient creates and returns a Client with the given public key and secret key.  Signs
// in to http://omise.co and visit https://dashboard.omise.co/test/dashboard to obtain
// your test (or live) keys.
func NewClient(pkey, skey string) (*Client, error) {
	switch {
	case pkey == "" && skey == "":
		return nil, ErrInvalidKey
	case pkey != "" && !strings.HasPrefix(pkey, "pkey_"):
		return nil, ErrInvalidKey
	case skey != "" && !strings.HasPrefix(skey, "skey_"):
		return nil, ErrInvalidKey
	}

	client := &Client{
		Client: &http.Client{Transport: transport},
		debug:  false,
		pkey:   pkey,
		skey:   skey,

		Endpoints: map[internal.Endpoint]string{},
	}

	if len(build.Default.ReleaseTags) > 0 {
		client.GoVersion = build.Default.ReleaseTags[len(build.Default.ReleaseTags)-1]
	}

	return client, nil
}

func NewClientWithChainKey(ckey string) (*Client, error) {
	switch {
	case ckey == "":
		return nil, ErrInvalidKey
	case ckey != "" && !strings.HasPrefix(ckey, "ckey_"):
		return nil, ErrInvalidKey
	}

	client := &Client{
		Client: &http.Client{Transport: transport},
		debug:  false,
		ckey:   ckey,

		Endpoints: map[internal.Endpoint]string{},
	}
	if len(build.Default.ReleaseTags) > 0 {
		client.GoVersion = build.Default.ReleaseTags[len(build.Default.ReleaseTags)-1]
	}
	return client, nil
}

// Request creates a new *http.Request that should performs the supplied Operation. Most
// people should use the Do method instead.
func (c *Client) Request(operation internal.Operation) (req *http.Request, err error) {
	req, err = c.buildJSONRequest(operation)
	if err != nil {
		return nil, err
	}

	err = c.setRequestHeaders(req, operation.Describe())
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) buildJSONRequest(operation internal.Operation) (*http.Request, error) {
	desc := operation.Describe()

	b, err := json.Marshal(operation)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(b)

	endpoint := string(desc.Endpoint)
	if ep, ok := c.Endpoints[desc.Endpoint]; ok {
		endpoint = ep
	}

	return http.NewRequest(desc.Method, endpoint+desc.Path, body)
}

func (c *Client) setRequestHeaders(req *http.Request, desc *internal.Description) error {
	ua := "OmiseGo/2015-11-06"
	if c.GoVersion != "" {
		ua += " Go/" + c.GoVersion
	}

	if desc.ContentType != "" {
		req.Header.Add("Content-Type", desc.ContentType)
	}

	req.Header.Add("User-Agent", ua)
	if c.APIVersion != "" {
		req.Header.Add("Omise-Version", c.APIVersion)
	}

	switch desc.KeyKind() {
	case "public":
		req.SetBasicAuth(c.pkey, "")
	case "secret":
		switch {
		case c.skey != "":
			req.SetBasicAuth(c.skey, "")
		case c.ckey != "":
			req.SetBasicAuth(c.ckey, "")
		default:
			return ErrInternal("either secret key or chain key must not be empty")
		}
	default:
		return ErrInternal("unrecognized endpoint:" + desc.Endpoint)
	}

	return nil
}

// Do performs the supplied operation against Omise's REST API and unmarshal the response
// into the given result parameter. Results are usually basic objects or a list that
// corresponds to the operations being done.
//
// If the operation is successful, result should contains the response data. Otherwise a
// non-nil error should be returned. Error maybe of the omise-go.Error struct type, in
// which case you can further inspect the Code and Message field for more information.
func (c *Client) Do(result interface{}, operation internal.Operation) error {
	req, err := c.Request(operation)
	if err != nil {
		return err
	}

	// response
	resp, err := c.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrTransport{err, buffer}
	}

	switch {
	case resp.StatusCode != 200:
		err := &Error{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(buffer, err); err != nil {
			return &ErrTransport{err, buffer}
		}

		return err
	} // status == 200 && e == nil

	if c.debug {
		fmt.Println("resp:", resp.StatusCode, string(buffer))
	}

	if result != nil {
		if err := json.Unmarshal(buffer, result); err != nil {
			return &ErrTransport{err, buffer}
		}
	}

	return nil
}

func (c *Client) FormDataRequest(operation internal.Operation) (req *http.Request, err error) {
	req, err = c.buildFormDataRequest(operation)
	if err != nil {
		return nil, err
	}

	err = c.setRequestHeaders(req, operation.Describe())
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) buildFormDataRequest(operation internal.Operation) (*http.Request, error) {
	desc := operation.Describe()

	form := url.Values{}
	err := encoder.Encode(operation, form)
	if err != nil {
		return nil, err
	}

	body := strings.NewReader(form.Encode())

	endpoint := string(desc.Endpoint)
	if ep, ok := c.Endpoints[desc.Endpoint]; ok {
		endpoint = ep
	}

	return http.NewRequest(desc.Method, endpoint+desc.Path, body)
}

func (c *Client) DoWithFormData(result interface{}, operation internal.Operation) error {
	req, err := c.FormDataRequest(operation)
	if err != nil {
		return err
	}

	// response
	resp, err := c.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrTransport{err, buffer}
	}

	switch {
	case resp.StatusCode != 200:
		err := &Error{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(buffer, err); err != nil {
			return &ErrTransport{err, buffer}
		}

		return err
	} // status == 200 && e == nil

	if c.debug {
		fmt.Println("resp:", resp.StatusCode, string(buffer))
	}

	if result != nil {
		if err := json.Unmarshal(buffer, result); err != nil {
			return &ErrTransport{err, buffer}
		}
	}

	return nil
}

func (c *Client) UploadDocumentRequest(operation internal.Operation) (req *http.Request, err error) {
	req, err = c.buildUploadDocumentRequest(operation)
	if err != nil {
		return nil, err
	}

	err = c.setRequestHeaders(req, operation.Describe())
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) buildUploadDocumentRequest(operation internal.Operation) (*http.Request, error) {
	desc := operation.Describe()

	b, err := json.Marshal(operation)
	if err != nil {
		return nil, err
	}

	document := struct {
		File     []byte
		Filename string
		Kind     string
	}{}

	if err := json.Unmarshal(b, &document); err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("kind", document.Kind); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("file", document.Filename)
	if err != nil {
		return nil, err
	}

	file := bytes.NewReader(document.File)

	io.Copy(part, file)
	writer.Close()

	endpoint := string(desc.Endpoint)
	if ep, ok := c.Endpoints[desc.Endpoint]; ok {
		endpoint = ep
	}

	req, err := http.NewRequest(desc.Method, endpoint+desc.Path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	return req, err
}

func (c *Client) DoUploadDocument(result interface{}, operation internal.Operation) error {
	req, err := c.UploadDocumentRequest(operation)
	if err != nil {
		return err
	}

	// response
	resp, err := c.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrTransport{err, buffer}
	}

	switch {
	case resp.StatusCode != 200:
		err := &Error{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(buffer, err); err != nil {
			return &ErrTransport{err, buffer}
		}

		return err
	} // status == 200 && e == nil

	if c.debug {
		fmt.Println("resp:", resp.StatusCode, string(buffer))
	}

	if result != nil {
		if err := json.Unmarshal(buffer, result); err != nil {
			return &ErrTransport{err, buffer}
		}
	}
	return nil
}
