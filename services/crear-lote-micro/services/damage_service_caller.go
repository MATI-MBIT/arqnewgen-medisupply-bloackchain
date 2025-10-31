package services

import (
	"CrearLoteMicro/models"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type DamageServiceCaller struct {
	client  *http.Client
	postURL string
}

func NewDamageServiceCaller(url string) *DamageServiceCaller {
	return &DamageServiceCaller{
		client:  &http.Client{Timeout: 5 * time.Second},
		postURL: url,
	}
}

func (d *DamageServiceCaller) SendLoteInfo(info *models.LoteInfoResponse) error {
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, d.postURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("DamageServiceCaller non-2xx response: %s", resp.Status)
	}
	return nil
}
