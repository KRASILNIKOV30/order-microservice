package transport

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKitty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/kitty/dora?age=25", nil)
	w := httptest.NewRecorder()
	router := Router()
	router.ServeHTTP(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Status code is wrong. Have: %d, want: %d.", res.StatusCode, http.StatusOK)
	}

	jsonString, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	var kitty Kitty

	if err = json.Unmarshal(jsonString, &kitty); err != nil {
		t.Errorf("Can't parse json response with error %v", err)
	}

	expected := Kitty{"dora25"}
	if kitty != expected {
		t.Errorf("Unexpected kitty returned. Have: %v, want: %v", kitty, expected)
	}
}
