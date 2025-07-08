package httpclient

import "testing"

func TestGet(t *testing.T) {
	cmd := "http://localhost:8089/ping"
	result, err := Get(cmd, 30)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestPostJson(t *testing.T) {
	url := "http://localhost:8089/some/json/endpoint"

	jsonBody := `{"message": "hello", "user": "test"}`

	result, err := PostJson(url, jsonBody, 30)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)

}
