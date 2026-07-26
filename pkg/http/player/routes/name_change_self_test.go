package routes

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
)

// TestSelfNameChangeChecksAndConfirms verifies the browser-compatible flow.
func TestSelfNameChangeChecksAndConfirms(t *testing.T) {
	app, manager, _ := nameChangeApplication(t, false)

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
		manager.record.Profile.LastNameChangeAt == nil ||
		result.AvailableAt == nil {
		t.Fatalf(
			"confirm status=%d result=%#v record=%#v",
			response.StatusCode,
			result,
			manager.record,
		)
	}
}

// TestSelfNameChangeReportsCooldown verifies policy parity with Nitro.
func TestSelfNameChangeReportsCooldown(t *testing.T) {
	app, manager, _ := nameChangeApplication(t, false)
	changedAt := time.Now().UTC()
	manager.record.Profile.LastNameChangeAt = &changedAt
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
		result.Code != playeridentity.ResultCooldown ||
		result.AvailableAt == nil {
		t.Fatalf("status=%d result=%#v", response.StatusCode, result)
	}
}
