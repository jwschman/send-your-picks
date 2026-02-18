package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Odds API structures
type OddsOutcome struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Point float64 `json:"point"`
}

type OddsMarket struct {
	Key        string        `json:"key"`
	LastUpdate time.Time     `json:"last_update"`
	Outcomes   []OddsOutcome `json:"outcomes"`
}

type OddsBookmaker struct {
	Key        string       `json:"key"`
	Title      string       `json:"title"`
	LastUpdate time.Time    `json:"last_update"`
	Markets    []OddsMarket `json:"markets"`
}

type OddsGame struct {
	ID           string          `json:"id"`
	SportKey     string          `json:"sport_key"`
	SportTitle   string          `json:"sport_title"`
	CommenceTime time.Time       `json:"commence_time"`
	HomeTeam     string          `json:"home_team"`
	AwayTeam     string          `json:"away_team"`
	Bookmakers   []OddsBookmaker `json:"bookmakers"`
}

// SpreadInfo contains the spread for a game
type SpreadInfo struct {
	HomeTeamAbbr string
	AwayTeamAbbr string
	HomeSpread   float64   // positive means home team is favored by this amount
	AwaySpread   float64   // negative of HomeSpread
	Bookmaker    string    // which bookmaker this spread is from
	LastUpdate   time.Time
	CommenceTime time.Time // when the game starts
}

// OddsClient for The Odds API
type OddsClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// Team name mapping from Odds API format to your internal abbreviations
var oddsAPIToAbbreviation = map[string]string{
	"Arizona Cardinals":     "ARI",
	"Atlanta Falcons":       "ATL",
	"Baltimore Ravens":      "BAL",
	"Buffalo Bills":         "BUF",
	"Carolina Panthers":     "CAR",
	"Chicago Bears":         "CHI",
	"Cincinnati Bengals":    "CIN",
	"Cleveland Browns":      "CLE",
	"Dallas Cowboys":        "DAL",
	"Denver Broncos":        "DEN",
	"Detroit Lions":         "DET",
	"Green Bay Packers":     "GB",
	"Houston Texans":        "HOU",
	"Indianapolis Colts":    "IND",
	"Jacksonville Jaguars":  "JAX",
	"Kansas City Chiefs":    "KC",
	"Las Vegas Raiders":     "LV",
	"Los Angeles Chargers":  "LAC",
	"Los Angeles Rams":      "LAR",
	"Miami Dolphins":        "MIA",
	"Minnesota Vikings":     "MIN",
	"New England Patriots":  "NE",
	"New Orleans Saints":    "NO",
	"New York Giants":       "NYG",
	"New York Jets":         "NYJ",
	"Philadelphia Eagles":   "PHI",
	"Pittsburgh Steelers":   "PIT",
	"San Francisco 49ers":   "SF",
	"Seattle Seahawks":      "SEA",
	"Tampa Bay Buccaneers":  "TB",
	"Tennessee Titans":      "TEN",
	"Washington Commanders": "WSH",
}

// NewOddsClient creates a new Odds API client
func NewOddsClient() (*OddsClient, error) {
	apiKey := os.Getenv("ODDS_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ODDS_API_KEY not set")
	}

	// Allow override of base URL for testing with mock server
	baseURL := os.Getenv("ODDS_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.the-odds-api.com/v4"
	}

	return &OddsClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// FetchSpreads fetches current spreads for NFL games
// preferredBookmaker can be empty string to use the first available bookmaker
// Common bookmakers: "draftkings", "fanduel", "betmgm", "caesars"
func (c *OddsClient) FetchSpreads(ctx context.Context, preferredBookmaker string) ([]SpreadInfo, error) {
	url := fmt.Sprintf("%s/sports/americanfootball_nfl/odds/?apiKey=%s&regions=us&markets=spreads",
		c.baseURL, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("odds API returned %s", resp.Status)
	}

	var games []OddsGame
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		return nil, err
	}

	var spreads []SpreadInfo
	for _, game := range games {
		// Convert team names to abbreviations
		homeAbbr, ok := oddsAPIToAbbreviation[game.HomeTeam]
		if !ok {
			return nil, fmt.Errorf("unknown team name: %s", game.HomeTeam)
		}
		awayAbbr, ok := oddsAPIToAbbreviation[game.AwayTeam]
		if !ok {
			return nil, fmt.Errorf("unknown team name: %s", game.AwayTeam)
		}

		// Find the preferred bookmaker or use the first available
		var selectedBookmaker *OddsBookmaker
		if preferredBookmaker != "" {
			for i := range game.Bookmakers {
				if game.Bookmakers[i].Key == preferredBookmaker {
					selectedBookmaker = &game.Bookmakers[i]
					break
				}
			}
		}
		if selectedBookmaker == nil && len(game.Bookmakers) > 0 {
			selectedBookmaker = &game.Bookmakers[0]
		}

		if selectedBookmaker == nil {
			continue // No bookmakers available for this game
		}

		// Extract spread from markets
		for _, market := range selectedBookmaker.Markets {
			if market.Key != "spreads" {
				continue
			}

			// Find home and away spreads
			var homeSpread, awaySpread float64
			for _, outcome := range market.Outcomes {
				if outcome.Name == game.HomeTeam {
					homeSpread = outcome.Point
				} else if outcome.Name == game.AwayTeam {
					awaySpread = outcome.Point
				}
			}

			spreads = append(spreads, SpreadInfo{
				HomeTeamAbbr: homeAbbr,
				AwayTeamAbbr: awayAbbr,
				HomeSpread:   homeSpread,
				AwaySpread:   awaySpread,
				Bookmaker:    selectedBookmaker.Title,
				LastUpdate:   selectedBookmaker.LastUpdate,
				CommenceTime: game.CommenceTime,
			})
		}
	}

	return spreads, nil
}

// GetTeamAbbreviation converts an Odds API team name to your internal abbreviation
func GetTeamAbbreviation(oddsAPITeamName string) (string, bool) {
	abbr, ok := oddsAPIToAbbreviation[oddsAPITeamName]
	return abbr, ok
}
