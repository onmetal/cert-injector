/*
Copyright (c) 2021 T-Systems International GmbH, SAP SE or an SAP affiliate company. All right reserved
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var envs = map[string]string{
	"PORT":        "8080",
	"DOMAIN_NAME": "domain.com",
	"TOKEN":       "123456",
	"AUTH_KEY":    "123456",
}

func TestServerRun(t *testing.T) {
	a := assert.New(t)
	if err := os.Setenv("PORT", "0x16"); err != nil {
		t.Logf("failed to setup env: %s", err)
		t.Fail()
	}

	server := New()
	err := server.Run()
	if !a.Contains(err.Error(), "unknown port") {
		t.Logf("server didn't run properly: %s", err)
		t.Fail()
	}
}

func TestHandler(t *testing.T) {
	a := assert.New(t)
	if err := setupEnvs(); err != nil {
		t.Logf("failed to setup env: %s", err)
		t.Fail()
	}
	server := New()
	go func() {
		err := server.Run()
		if err != nil {
			t.Logf("server didn't run properly: %s", err)
			t.Fail()
		}
	}()
	time.Sleep(100 * time.Millisecond)
	token := os.Getenv("TOKEN")
	if !assert.NotEmpty(t, token, "token env is empty") {
		t.Fatal()
	}
	path := fmt.Sprintf("%s/%s", HTTPChallengePath, token)

	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   path,
		},
		Host: "domain.com",
	}
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req) // nolint:bodyclose
	if err != nil {
		t.Logf("failed to do a request: %s", err)
		t.Fail()
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Logf("failed to close response body: %s", err)
			t.Fail()
		}
	}(resp.Body)
	a.Equal(http.StatusOK, resp.StatusCode, "expected 200")
}

func setupEnvs() error {
	for name, value := range envs {
		if err := os.Setenv(name, value); err != nil {
			return err
		}
	}
	return nil
}
