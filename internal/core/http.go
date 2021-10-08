// Copyright © 2021 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hyperledger/firefly-cli/internal/log"
)

type HttpClient struct {
	log log.Logger
}

func NewHttpClient(log log.Logger) *HttpClient {
	return &HttpClient{
		log: log,
	}
}

func (h *HttpClient) RequestWithRetry(method, url string, body, result interface{}) (err error) {
	retries := 30
	for {
		if err := request(method, url, body, result); err != nil {
			if retries > 0 {
				h.log.Warnf("%s - retrying request...", err.Error())
				retries--
				time.Sleep(1 * time.Second)
			} else {
				return err
			}
		} else {
			return nil
		}
	}
}

func request(method, url string, body, result interface{}) (err error) {
	if body == nil {
		body = make(map[string]interface{})
	}

	var bodyReader io.Reader
	if body != nil {
		requestBody, err := json.Marshal(&body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(requestBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var responseBytes []byte
		if resp.StatusCode != 204 {
			responseBytes, _ = ioutil.ReadAll(resp.Body)
		}
		return fmt.Errorf("%s returned %d: %s", url, resp.StatusCode, responseBytes)
	}

	if resp.StatusCode == 204 {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(&result)
}
