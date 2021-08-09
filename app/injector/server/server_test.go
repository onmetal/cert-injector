package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const waitForServerStartMillisecond = 100 * time.Millisecond

var deployment = []byte(`{
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
        "name": "test",
        "creationTimestamp": null,
		"annotations": {
			"cert.injector.ko/mount": "true",
			"cert.injector.ko/cert-name": "test-cert"
		},
        "labels": {
            "app": "test"
        }
    },
    "spec": {
        "replicas": 1,
        "selector": {
            "matchLabels": {
                "app": "test"
            }
        },
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "app": "test"
                }
            },
            "spec": {
                "containers": [
                    {
                        "name": "nginx",
                        "image": "nginx",
                        "resources": {}
                    }
                ]
            }
        },
        "strategy": {}
    },
    "status": {}
}
`)

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("can't find file")
		os.Exit(1)
	}
	dir := path.Join(path.Dir(filename), "..", "..", "..")
	err := os.Chdir(dir)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func TestServerRun(t *testing.T) {
	setupEnvs(t)
	server := New()
	go func() {
		err := server.Run()
		if err != nil {
			t.Logf("server didn't run properly: %s", err)
			t.Fail()
		}
	}()
	time.Sleep(waitForServerStartMillisecond)
}

func TestEmptyBody(t *testing.T) {
	a := assert.New(t)
	client := &http.Client{Timeout: 30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}} // nolint:gosec
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost:8443",
			Path:   "/api/v1/mutate",
		},
		Body: nil,
	}
	resp, err := client.Do(req) // nolint:bodyclose
	if err != nil {
		t.Logf("failed to do a request: %s", err)
		t.Fail()
	}
	defer func(Body io.ReadCloser) {
		bodyErr := Body.Close()
		if bodyErr != nil {
			t.Logf("failed to close response body: %s", bodyErr)
			t.Fail()
		}
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("failed read response body: %s", err)
		t.Fail()
	}

	a.Equal(http.StatusBadRequest, resp.StatusCode, "expected 400")
	a.Contains(string(body), "empty body")
}

func TestApplicationType(t *testing.T) {
	a := assert.New(t)
	var jsonStr = []byte(`{
		"uid":"1234",
		"kind": {"group": "apps", "version": "v1", "kind": "Deployment"}
	}`)
	client := &http.Client{Timeout: 30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}} // nolint:gosec

	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost:8443",
			Path:   "/api/v1/mutate",
		},
		Body: ioutil.NopCloser(bytes.NewBuffer(jsonStr)),
	}
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

	a.Equal(http.StatusUnsupportedMediaType, resp.StatusCode, "expected 415")
}

func TestDecode(t *testing.T) {
	a := assert.New(t)
	var jsonStr = []byte(`{
		"uid":"1234",
		"kind": {"group": "apps", "version": "v1", "kind": "Deployment"}
	}`)
	client := &http.Client{Timeout: 30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}} // nolint:gosec
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost:8443",
			Path:   "/api/v1/mutate",
		},
		Header: map[string][]string{"Content-Type": {"application/json"}},
		Body:   ioutil.NopCloser(bytes.NewBuffer(jsonStr)),
	}
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

func TestMutate(t *testing.T) {
	a := assert.New(t)
	admissionRequest := &v1.AdmissionReview{
		Request: &v1.AdmissionRequest{
			UID: "1234",
			Kind: metav1.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "deployment",
			},
			Object: pkgruntime.RawExtension{Raw: deployment},
		},
	}
	requestBody, err := json.Marshal(admissionRequest)
	if err != nil {
		t.Logf("failed marshal requestBody request: %s", err)
		t.Fail()
	}
	client := &http.Client{Timeout: 30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}} //nolint:gosec
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost:8443",
			Path:   "/api/v1/mutate",
		},
		Header: map[string][]string{"Content-Type": {"application/json"}},
		Body:   ioutil.NopCloser(bytes.NewBuffer(requestBody)),
	}
	resp, err := client.Do(req) // nolint:bodyclose
	if err != nil {
		t.Logf("failed to do a request: %s", err)
		t.Fail()
	}
	defer func(Body io.ReadCloser) {
		bodyErr := Body.Close()
		if bodyErr != nil {
			t.Logf("failed to close response requestBody: %s", bodyErr)
			t.Fail()
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("failed read response body: %s", err)
		t.Fail()
	}
	admissionResponse := &v1.AdmissionReview{}
	if _, _, err := deserializer.Decode(respBody, nil, admissionResponse); err != nil {
		t.Logf("Can't decode body: %s", err)
		t.Fail()
	}
	a.Equal(http.StatusOK, resp.StatusCode, "expected 200")
	a.Equal(types.UID("1234"), admissionResponse.Response.UID)
}

func setupEnvs(t *testing.T) {
	/* self-signed cert might be created by openssl command:
	openssl req -nodes -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 10000 -subj '/CN=localhost' */
	if err := os.Setenv("CERT_PATH", "certs/cert.pem"); err != nil {
		t.Logf("failed to setup env: %s", err)
		t.Fail()
	}
	if err := os.Setenv("KEY_PATH", "certs/key.pem"); err != nil {
		t.Logf("failed to setup env: %s", err)
		t.Fail()
	}
}
