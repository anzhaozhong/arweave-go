package arweave

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ArweaveClient struct {
	client *http.Client
	url    string
}

func Dial(ctx context.Context, rawurl string) (*ArweaveClient, error) {
	_, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, rawurl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return &ArweaveClient{client: new(http.Client), url: rawurl}, nil
}

func (c *ArweaveClient) GetData(txID string) (string, error) {
	body, err := c.get(fmt.Sprintf("tx/%s/data", txID))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *ArweaveClient) LastTransaction(address string) (string, error) {
	body, err := c.get(fmt.Sprintf("wallet/%s/last_tx", address))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *ArweaveClient) GetTransaction(txID string) (*JsonTransaction, error) {
	body, err := c.get(fmt.Sprintf("tx/%s", txID))
	if err != nil {
		return nil, err
	}
	tx := JsonTransaction{}
	err = json.Unmarshal(body, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

var allowedFields = map[string]bool{
	"id":        true,
	"last_tx":   true,
	"owner":     true,
	"target":    true,
	"quantity":  true,
	"type":      true,
	"data":      true,
	"reward":    true,
	"signature": true,
	"data.html": true,
}

func (c *ArweaveClient) GetTransactionField(txID string, field string) (string, error) {
	_, ok := allowedFields[field]
	if !ok {
		return "", errors.New("field does not exist")
	}
	body, err := c.get(fmt.Sprintf("tx/%s/%s", txID, field))
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *ArweaveClient) GetBlockByID(blockID string) (*Block, error) {
	body, err := c.get(fmt.Sprintf("block/hash/%s", blockID))
	if err != nil {
		return nil, err
	}
	block := Block{}
	err = json.Unmarshal(body, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (c *ArweaveClient) GetBlockByHeight(height int64) (*Block, error) {
	body, err := c.get(fmt.Sprintf("block/height/%d", height))
	if err != nil {
		return nil, err
	}
	block := Block{}
	err = json.Unmarshal(body, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (c *ArweaveClient) GetCurrentBlock() (*Block, error) {
	body, err := c.get("current_block")
	if err != nil {
		return nil, err
	}
	block := Block{}
	err = json.Unmarshal(body, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (c *ArweaveClient) GetReward(data []byte) (string, error) {
	body, err := c.get(fmt.Sprintf("price/%d", len(data)))
	if err != nil {
		return "", err
	}
	return string(body), nil

}
func (c *ArweaveClient) GetBalance(address string) (string, error) {
	body, err := c.get(fmt.Sprintf("wallet/%s/balance", address))
	if err != nil {
		return "", err
	}
	return string(body), nil

}

func (c *ArweaveClient) GetPeers() ([]string, error) {
	body, err := c.get("peers")
	if err != nil {
		return nil, err
	}
	peers := []string{}
	err = json.Unmarshal(body, &peers)
	if err != nil {
		return nil, err
	}

	return peers, nil

}

func (c *ArweaveClient) GetInfo() (*NetworkInfo, error) {
	body, err := c.get("info")
	if err != nil {
		return nil, err
	}
	info := NetworkInfo{}
	json.Unmarshal(body, &info)
	return &info, nil
}

func (c *ArweaveClient) Commit(data []byte) (string, error) {
	body, err := c.post(context.TODO(), "tx", data)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *ArweaveClient) get(endpoint string) ([]byte, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/%s", c.url, endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("not found")
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error with body %s", string(b))
	}
	return b, err
}

func (c *ArweaveClient) post(ctx context.Context, endpoint string, body []byte) ([]byte, error) {
	r := bytes.NewReader(body)
	resp, err := c.client.Post(fmt.Sprintf("%s/%s", c.url, endpoint), "application/json", r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}