package routes

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
)

// TestSelfNameChangeChecksAndConfirms verifies the browser-compatible flow.
func TestSelfNameChangeChecksAndConfirms(t *testing.T) {
	app, manager, _ := nameChangeApplication(t, false)
	manager.record.Profile.AllowNameChange = true

	response := nameChangeRequest(
		t,
		app,
		http.MethodPost,
		"/api/admin/players/7/name-change/check",
		`{"username":"browser"}`,
	)
	var check NameCheckResponse
	if err := json.NewDecoder(response.Body).Decode(&check); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK ||
		check.Code != playeridentity.ResultAvailable {
		t.Fatalf("check status=%d result=%#v", response.StatusCode, check)
	}

	response = nameChangeRequest(
		t,
		app,
		http.MethodPost,
		"/api/admin/players/7/name-change/confirm",
		`{"username":"browser"}`,
	)
	var result NameCheckResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK ||
		result.Code != playeridentity.ResultAvailable ||
		manager.record.Player.Username != "browser" ||
		manager.record.Profile.AllowNameChange {
		t.Fatalf(
			"confirm status=%d result=%#v record=%#v",
			response.StatusCode,
			result,
			manager.record,
		)
	}
}

// TestSelfNameChangeReportsDisabled verifies policy parity with Nitro.
func TestSelfNameChangeReportsDisabled(t *testing.T) {
	app, _, _ := nameChangeApplication(t, false)
	response := nameChangeRequest(
		t,
		app,
		http.MethodPost,
		"/api/admin/players/7/name-change/check",
		`{"username":"browser"}`,
	)
	var result NameCheckResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK ||
		result.Code != playeridentity.ResultDisabled {
		t.Fatalf("status=%d result=%#v", response.StatusCode, result)
	}
}
