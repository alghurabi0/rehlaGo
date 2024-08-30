package fileStorage

import (
	"fmt"
	"io"
	"net/http"
)

type WistiaModel struct {
	Token string
}

func (w *WistiaModel) DeleteVideo(hashedId string) error {
	url := fmt.Sprintf("https://api.wistia.com/v1/medias/%s.json", hashedId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Printf("couldn't get new req: %v\n", err)
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.Token))
	req.Header.Set("Accept", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to send a req: %v\n", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Printf("unexpected status code: %d\n", res.StatusCode)
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("response body: %s\n", body)
		return fmt.Errorf("received non-2xx response code: %d", res.StatusCode)
	}
	return nil
}
