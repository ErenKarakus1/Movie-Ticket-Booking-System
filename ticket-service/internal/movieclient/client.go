package movieclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrShowtimeNotFound = errors.New("showtime not found")
	ErrUnavailable      = errors.New("movie service unavailable")
)

type Seat struct {
	ID uuid.UUID `json:"id"`
}

type Showtime struct {
	ID       uuid.UUID `json:"id"`
	StartsAt time.Time `json:"starts_at"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("movie service URL is required")
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}, nil
}

func (c *Client) GetShowtimeSeats(ctx context.Context, showtimeID uuid.UUID) ([]Seat, error) {
	url := fmt.Sprintf("%s/showtimes/%s/seats", c.baseURL, showtimeID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, ErrUnavailable
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrShowtimeNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnavailable
	}

	var seats []Seat
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&seats); err != nil {
		return nil, ErrUnavailable
	}
	return seats, nil
}

func (c *Client) GetShowtime(ctx context.Context, showtimeID uuid.UUID) (Showtime, error) {
	url := fmt.Sprintf("%s/showtimes/%s", c.baseURL, showtimeID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Showtime{}, ErrUnavailable
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Showtime{}, ErrUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Showtime{}, ErrShowtimeNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return Showtime{}, ErrUnavailable
	}

	var showtime Showtime
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&showtime); err != nil || showtime.StartsAt.IsZero() {
		return Showtime{}, ErrUnavailable
	}
	return showtime, nil
}
