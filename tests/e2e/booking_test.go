package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type response struct {
	status int
	body   []byte
}

func TestBookingFlowAndRejectedRequests(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("E2E_GATEWAY_URL"), "/")
	databaseURL := os.Getenv("E2E_DATABASE_URL")
	if baseURL == "" || databaseURL == "" {
		t.Skip("set E2E_GATEWAY_URL and E2E_DATABASE_URL to run the integration test")
	}

	db, err := pgx.Connect(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("connect to E2E database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close(context.Background()) })

	client := &http.Client{Timeout: 10 * time.Second}
	request := func(method, path, token string, body any) response {
		t.Helper()
		var payload []byte
		if body != nil {
			var err error
			payload, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s %s failed: %v", method, path, err)
		}
		defer res.Body.Close()
		var data bytes.Buffer
		if _, err := data.ReadFrom(res.Body); err != nil {
			t.Fatal(err)
		}
		return response{status: res.StatusCode, body: data.Bytes()}
	}
	assertStatus := func(got response, want int, operation string) {
		t.Helper()
		if got.status != want {
			t.Fatalf("%s: status=%d, want=%d; body=%s", operation, got.status, want, got.body)
		}
	}
	decode := func(data []byte, dst any) {
		t.Helper()
		if err := json.Unmarshal(data, dst); err != nil {
			t.Fatalf("decode response %s: %v", data, err)
		}
	}

	// Rejected paths: no token, invalid IDs, invalid request data, and duplicate seat holds.
	assertStatus(request(http.MethodPost, "/auditoriums", "", map[string]string{"name": "forbidden"}), http.StatusUnauthorized, "admin endpoint without token")
	assertStatus(request(http.MethodGet, "/movies/not-a-uuid", "", nil), http.StatusBadRequest, "movie endpoint with invalid ID")
	assertStatus(request(http.MethodPost, "/auth/register", "", map[string]string{"name": "", "email": "invalid", "password": "short"}), http.StatusBadRequest, "registration with invalid data")

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	adminEmail := "e2e-admin-" + suffix + "@example.com"
	adminPassword := "ValidAdminPass123"
	var movieID, auditoriumID, showtimeID, bookingID, testAdminID string
	var testUserIDs []string
	var seatIDs []string
	var holdID, userToken string
	t.Cleanup(func() {
		ctx := context.Background()
		if holdID != "" {
			request(http.MethodDelete, "/holds/"+holdID, userToken, nil)
		}
		if bookingID != "" {
			_, _ = db.Exec(ctx, "DELETE FROM booking_seats WHERE booking_id = $1", bookingID)
			_, _ = db.Exec(ctx, "DELETE FROM bookings WHERE id = $1", bookingID)
		}
		if showtimeID != "" {
			_, _ = db.Exec(ctx, "DELETE FROM showtimes WHERE id = $1", showtimeID)
		}
		if len(seatIDs) > 0 {
			for _, seatID := range seatIDs {
				_, _ = db.Exec(ctx, "DELETE FROM seats WHERE id = $1", seatID)
			}
		}
		if auditoriumID != "" {
			_, _ = db.Exec(ctx, "DELETE FROM auditoriums WHERE id = $1", auditoriumID)
		}
		if movieID != "" {
			_, _ = db.Exec(ctx, "DELETE FROM movies WHERE id = $1", movieID)
		}
		for _, testUserID := range testUserIDs {
			_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", testUserID)
		}
		if testAdminID != "" {
			_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", testAdminID)
		}
	})

	adminRegistration := request(http.MethodPost, "/auth/register", "", map[string]string{"name": "E2E Admin", "email": adminEmail, "password": adminPassword})
	assertStatus(adminRegistration, http.StatusCreated, "register temporary admin")
	var adminUser struct {
		ID string `json:"id"`
	}
	decode(adminRegistration.body, &adminUser)
	if adminUser.ID == "" {
		t.Fatalf("admin registration response has no user id: %s", adminRegistration.body)
	}
	testAdminID = adminUser.ID
	if _, err := db.Exec(context.Background(), "UPDATE users SET role = 'admin' WHERE id = $1", testAdminID); err != nil {
		t.Fatalf("promote temporary E2E user in test database: %v", err)
	}
	adminLogin := request(http.MethodPost, "/auth/login", "", map[string]string{"email": adminEmail, "password": adminPassword})
	assertStatus(adminLogin, http.StatusOK, "log in temporary admin")
	var loginBody struct {
		Token string `json:"token"`
	}
	decode(adminLogin.body, &loginBody)
	if loginBody.Token == "" {
		t.Fatal("admin login returned an empty token")
	}
	assertStatus(request(http.MethodPost, "/auth/login", "", map[string]string{"email": adminEmail, "password": "wrong-password"}), http.StatusBadRequest, "login with wrong password")
	var movie struct {
		ID string `json:"id"`
	}
	created := request(http.MethodPost, "/movies", loginBody.Token, map[string]any{
		"title": "E2E " + suffix, "description": "Integration test screening", "duration_minutes": 90, "release_date": "2020-01-01",
	})
	assertStatus(created, http.StatusCreated, "create test movie")
	decode(created.body, &movie)
	if movie.ID == "" {
		t.Fatalf("movie response has no id: %s", created.body)
	}
	movieID = movie.ID

	var auditorium struct {
		ID string `json:"id"`
	}
	created = request(http.MethodPost, "/auditoriums", loginBody.Token, map[string]string{"name": "E2E " + suffix})
	assertStatus(created, http.StatusCreated, "create test auditorium")
	decode(created.body, &auditorium)
	if auditorium.ID == "" {
		t.Fatalf("auditorium response has no id: %s", created.body)
	}
	auditoriumID = auditorium.ID

	seats := request(http.MethodPost, "/auditoriums/"+auditorium.ID+"/seats", loginBody.Token, []map[string]any{{"row_label": "E2E" + suffix, "seat_number": 1}})
	assertStatus(seats, http.StatusCreated, "create test seat")
	var seatRows []struct {
		ID string `json:"id"`
	}
	decode(seats.body, &seatRows)
	if len(seatRows) != 1 || seatRows[0].ID == "" {
		t.Fatalf("unexpected seat response: %s", seats.body)
	}
	seatIDs = []string{seatRows[0].ID}

	startsAt := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
	showtimeBody := map[string]any{"auditorium_id": auditorium.ID, "starts_at": startsAt.Format(time.RFC3339), "price_cents": 1200}
	showtime := request(http.MethodPost, "/movies/"+movie.ID+"/showtimes", loginBody.Token, showtimeBody)
	assertStatus(showtime, http.StatusCreated, "create test showtime")
	var showtimeData struct {
		ID string `json:"id"`
	}
	decode(showtime.body, &showtimeData)
	if showtimeData.ID == "" {
		t.Fatalf("showtime response has no id: %s", showtime.body)
	}
	showtimeID = showtimeData.ID

	// Scheduling another screening in the same auditorium at the same time must conflict.
	assertStatus(request(http.MethodPost, "/movies/"+movie.ID+"/showtimes", loginBody.Token, showtimeBody), http.StatusConflict, "overlapping showtime")

	var user struct {
		ID string `json:"id"`
	}
	registrationEmail := "e2e-" + suffix + "@example.com"
	registered := request(http.MethodPost, "/auth/register", "", map[string]string{"name": "E2E User", "email": registrationEmail, "password": "ValidPass123"})
	assertStatus(registered, http.StatusCreated, "register test user")
	decode(registered.body, &user)
	if user.ID == "" {
		t.Fatalf("registration response has no user id: %s", registered.body)
	}
	testUserIDs = append(testUserIDs, user.ID)

	userLogin := request(http.MethodPost, "/auth/login", "", map[string]string{"email": registrationEmail, "password": "ValidPass123"})
	assertStatus(userLogin, http.StatusOK, "log in test user")
	var userLoginBody struct {
		Token string `json:"token"`
	}
	decode(userLogin.body, &userLoginBody)
	userToken = userLoginBody.Token
	seatID := seatRows[0].ID
	hold := request(http.MethodPost, "/showtimes/"+showtimeData.ID+"/holds", userLoginBody.Token, map[string][]string{"seat_ids": {seatID}})
	assertStatus(hold, http.StatusCreated, "hold test seat")
	var holdData struct {
		HoldID string `json:"hold_id"`
	}
	decode(hold.body, &holdData)
	if holdData.HoldID == "" {
		t.Fatalf("hold response has no hold id: %s", hold.body)
	}
	holdID = holdData.HoldID
	secondEmail := "e2e-second-" + suffix + "@example.com"
	secondRegistration := request(http.MethodPost, "/auth/register", "", map[string]string{"name": "Second User", "email": secondEmail, "password": "ValidPass123"})
	assertStatus(secondRegistration, http.StatusCreated, "register second test user")
	var secondUser struct {
		ID string `json:"id"`
	}
	decode(secondRegistration.body, &secondUser)
	if secondUser.ID == "" {
		t.Fatalf("second registration response has no user id: %s", secondRegistration.body)
	}
	testUserIDs = append(testUserIDs, secondUser.ID)
	secondLogin := request(http.MethodPost, "/auth/login", "", map[string]string{"email": secondEmail, "password": "ValidPass123"})
	assertStatus(secondLogin, http.StatusOK, "log in second test user")
	var secondLoginBody struct {
		Token string `json:"token"`
	}
	decode(secondLogin.body, &secondLoginBody)
	// A second hold by another user must fail while the first user's hold is live.
	assertStatus(request(http.MethodPost, "/showtimes/"+showtimeData.ID+"/holds", secondLoginBody.Token, map[string][]string{"seat_ids": {seatID}}), http.StatusConflict, "hold already-held seat")

	booking := request(http.MethodPost, "/holds/"+holdData.HoldID+"/confirm", userLoginBody.Token, nil)
	assertStatus(booking, http.StatusCreated, "confirm test booking")
	var bookingData struct {
		ID string `json:"id"`
	}
	decode(booking.body, &bookingData)
	if bookingData.ID == "" {
		t.Fatalf("booking response has no id: %s", booking.body)
	}
	bookingID = bookingData.ID
	assertStatus(request(http.MethodGet, "/bookings/"+bookingData.ID, userLoginBody.Token, nil), http.StatusOK, "get confirmed booking")
}
