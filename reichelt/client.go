// The reichelt package is an implementation of a client for the Reichelt
// Elektronik search function to obtain product numbers. This is a dirty hack,
// don't use it. This is a fucking hack. It smells like shit, but it works.
// Like their online shop. But it's so shitty that I don't want to support this
// code, so therefore this code is hopefully in the public domain.

package reichelt

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	Endpoint string
	http.Client
}

type Record struct {
	ID int64
	Name string
}

var DefaultClient = Client{
	Endpoint: "http://www.reichelt.de/index.html",
}

func Search(term string) ([]Record, error) {
	return DefaultClient.Search(term)
}

func SearchID(name string) (int64, error) {
	return DefaultClient.SearchID(name)
}

func ResolveID(id string) string {
	res, err := SearchID(id)
	if err != nil {
		return id
	}
	return strconv.Itoa(int(res))
}

func (c *Client) Search(term string) ([]Record, error) {
	req, err := http.NewRequest("GET", c.Endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = url.Values{
		"ACTION": { "514" },
		"id": { "6" },
		"term": { term },
	}.Encode()

	res, err := c.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	rawRecords := []map[string]string{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&rawRecords)
	if err != nil {
		return nil, err
	}

	records := make([]Record, len(rawRecords))

	for n, rawRecord := range rawRecords {
		rawID := rawRecord["id"]
		id, err := strconv.Atoi(rawID[:len(rawID)-1])
		if err != nil {
			return nil, err
		}
		records[n] = Record{
			ID: int64(id),
			Name: rawRecord["value"],
		}
	}

	return records, nil
}

func (c *Client) SearchID(name string) (int64, error) {
	records, err := c.Search(name)
	if err != nil {
		return 0, err
	}
	for _, record := range records {
		if record.Name == name {
			return record.ID, nil
		}
	}
	return 0, nil
}
