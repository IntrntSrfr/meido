package backend

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type BackendService interface {
	// oooohhhh
}

type BotBackendService struct {
	BaseURL string
	Client  *http.Client
}

func NewBackendService() *BotBackendService {
	return &BotBackendService{
		BaseURL: "weed",
		Client:  &http.Client{Timeout: time.Second * 10},
	}
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type successResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func (c *BotBackendService) Request(method, endpoint string, data interface{}) ([]byte, error) {
	var body []byte
	var err error
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, nil
		}
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := resp.Body.Close()
		if err2 != nil {
			log.Println("error closing resp body")
		}
	}()

	return response, nil
}

func (c *BotBackendService) request() {

}
