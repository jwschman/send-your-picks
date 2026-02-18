// external API clients for game data and spreads
// if a provider ever disappears, only this package needs to change

package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// data from external API that i want to use
type ExternalTeam struct {
	Abbreviation string `json:"abbreviation"`
	Name         string `json:"name"`
	City         string `json:"city"`
}

type ExternalGame struct {
	ID            int64        `json:"id"`
	Season        int          `json:"season"`
	Week          int          `json:"week"`
	Date          time.Time    `json:"date"`
	HomeTeam      ExternalTeam `json:"home_team"`
	AwayTeam      ExternalTeam `json:"visitor_team"`
	Status        string       `json:"status"` // i'm only going to care about "Final"
	HomeTeamScore *int         `json:"home_team_score"`
	AwayTeamScore *int         `json:"visitor_team_score"`
	IsPostseason  bool         `json:"postseason"`
}

// API wrapper response
type gamesResponse struct {
	Data []ExternalGame `json:"data"`
	Meta struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"meta"`
}

// BallDontLieClient is the HTTP client for the BallDontLie NFL API
type BallDontLieClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// NewBallDontLieClient creates a new BallDontLie API client from environment variables
func NewBallDontLieClient() (*BallDontLieClient, error) {
	baseURL := os.Getenv("EXTERNAL_API_BASE_URL")
	apiKey := os.Getenv("EXTERNAL_API_KEY")

	if baseURL == "" {
		return nil, errors.New("EXTERNAL_API_BASE_URL not set")
	}
	if apiKey == "" {
		return nil, errors.New("EXTERNAL_API_KEY not set")
	}

	return &BallDontLieClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// FetchGames retrieves all games for a given season week from the BallDontLie API
func (c *BallDontLieClient) FetchGames(ctx context.Context, season int, week int, postseason bool) ([]ExternalGame, error) {

	u, err := url.Parse(c.baseURL + "/games")
	if err != nil {
		return nil, err
	}

	// Skip the Pro Bowl gap: API week 4 is the Pro Bowl, so postseason
	// weeks 4+ need to be shifted by 1 to get the correct API week.
	apiWeek := week
	if postseason && week >= 4 {
		apiWeek = week + 1
	}

	q := u.Query()
	q.Add("seasons[]", fmt.Sprint(season))
	q.Add("weeks[]", fmt.Sprint(apiWeek))
	if postseason {
		q.Add("postseason", "true") // only request postseason games
	} else {
		q.Add("postseason", "false") // only request regular season games
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API (balldontlie) returned %s", resp.Status)
	}

	var parsed gamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return parsed.Data, nil
}
