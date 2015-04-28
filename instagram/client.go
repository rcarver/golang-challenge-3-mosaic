package instagram

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

var INSTAGRAM_URL = "https://api.instagram.com/v1%s"
var INSTAGRAM_CLIENT_ID = "6b2ea4cc0093441fb38990045a855e2a"
var INSTAGRAM_SECRET = "ea785b48dd014eaeb4fd97c0a23d6ae5"

// Client makes requests to Instagram.
type Client struct {
	BaseURL string
	UrlSigner
}

// NewClient creates an initialized Client.
func NewClient() Client {
	return Client{
		BaseURL:   INSTAGRAM_URL,
		UrlSigner: clientSecretSigner{INSTAGRAM_CLIENT_ID, INSTAGRAM_SECRET},
	}
}

type MediaList struct {
	Media      []Media `json:"data"`
	Pagination `json:"pagination"`
}

type Search struct {
	Media      []Media `json:"data"`
	Pagination `json:"pagination"`
}

type Pagination struct {
	NextURL  string `json:"next_url"`
	MaxTagID string `json:"next_max_tag_id"`
}

// media is either a photo or video. If it's a video, it has both
// Images and Videos representations. If it's a photo, it only has Images
// representations.
type Media struct {
	Type   string          `json:"type"`
	Images map[string]*Rep `json:"images"`
	Videos map[string]*Rep `json:"videos"`
}

// IsPhoto tells you if this is a photo. If it's not, it's a video.
func (m Media) IsPhoto() bool {
	return m.Type == "image"
}

// StandardImage returns the standard resolution image representation.
func (m Media) StandardImage() *Rep {
	return m.Images["standard_resolution"]
}

// Rep is a JPG representation of an image or video, located at a URL and at a
// specific width and height.
type Rep struct {
	URL     string `json:"url"`
	Width   uint   `json:"width"`
	Height  uint   `json:"height"`
	fetched bool
	body    io.ReadCloser
	code    int
}

// Read implements io.Reader to fetch the JPG data.
func (r *Rep) Read(p []byte) (int, error) {
	if !r.fetched {
		r.fetched = true
		res, err := http.Get(r.URL)
		if err != nil {
			return 0, err
		}
		r.code = res.StatusCode
		if r.code == http.StatusOK {
			r.body = res.Body
		}
	}
	if r.body != nil {
		return r.body.Read(p)
	}
	return 0, fmt.Errorf("failed to fetch data, response was code %d", r.code)
}

// Image returns an image object from the JPG.
func (r *Rep) Image() (image.Image, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}
	return jpeg.Decode(&buf)
}

// Popular calls the Instagram Popular API and returns the data.
func (c Client) Popular() (*MediaList, error) {
	var m MediaList
	params := map[string]string{
		"count": "100",
	}
	url := c.formatURL("/media/popular", params)
	err := c.getJSON(url, &m)
	return &m, err
}

// Search calls the Instagram Search API and returns the data.
func (c Client) Search(lat, lng string) (*MediaList, error) {
	var m MediaList
	params := map[string]string{
		"lat":   lat,
		"lng":   lng,
		"count": "100",
	}
	url := c.formatURL("/media/search", params)
	err := c.getJSON(url, &m)
	return &m, err
}

// Tagged calls the Instagram Tagged API and returns the data.
func (c Client) Tagged(tag, maxTagId string) (*MediaList, error) {
	var m MediaList
	params := map[string]string{
		"count":      "100",
		"max_tag_id": maxTagId,
	}
	endpoint := fmt.Sprintf("/tags/%s/media/recent", tag)
	url := c.formatURL(endpoint, params)
	err := c.getJSON(url, &m)
	return &m, err
}

// MediaList returns a MediaList from a URL. This can be used with pagination
// NextURL to make repeated calls to an API.
func (c Client) MediaList(url string) (*MediaList, error) {
	var m MediaList
	err := c.getJSON(url, &m)
	return &m, err
}

// getJSON calls a URL and marshals the resulting JSON into the data struct. If
// the response is anything but 200 an error is returned.
func (c Client) getJSON(url string, data interface{}) error {
	println(url)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Call to %s failed, status %d", url, res.StatusCode)
	}
	// TODO use streaming unmarshal
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	return nil
}

// formatURL combines query parameters to an endpoint, then signs the URL.
func (c Client) formatURL(endpoint string, params map[string]string) string {
	u, err := url.Parse(fmt.Sprintf(c.BaseURL, endpoint))
	if err != nil {
		panic("failed to parse instagram base url")
	}
	// Add custom params to the query string.
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}

	// Sign the URL.
	c.UrlSigner.Sign(endpoint, &q)

	// Set new query string and stringify.
	u.RawQuery = q.Encode()
	return u.String()
}

// UrlSigner is the interface for things that modify a query string for
// security purposes.
type UrlSigner interface {
	Sign(endpoint string, queryString *url.Values)
}

// clientSecretSigner implements UrlSigner, it adds the client_id and sig
// params to a query string.
type clientSecretSigner struct {
	ClientID string
	Secret   string
}

// Sign implements UrlSigner.
func (s clientSecretSigner) Sign(endpoint string, q *url.Values) {
	q.Set("client_id", s.ClientID)
	q.Set("sig", sig(s.Secret, endpoint, *q))
}

// sig calculates the signature for an Instagram URL.
// https://instagram.com/developer/secure-api-requests/
func sig(secret, endpoint string, params url.Values) string {
	// Parts is made up of the endpoint and params sorted by key.
	parts := make([]string, 0, len(params)+1)
	parts = append(parts, endpoint)

	// Get the sorted keys.
	keys := make([]string, 0, len(params))
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))

	// Accumulate sorted key/value pairs.
	for _, k := range keys {
		// TODO: use url.Values#Get?
		join := fmt.Sprintf("%s=%s", k, params[k][0])
		parts = append(parts, join)
	}

	// Join parts and sign it.
	sig := strings.Join(parts, "|")

	// Calculate the sha256 hexdigest.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sig))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum)
}
