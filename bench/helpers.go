// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/topfreegames/khan/util"
)

func getRoute(url string) string {
	return fmt.Sprintf("http://localhost:8888%s", url)
}

func postTo(url string, payload util.JSON) (*http.Response, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
