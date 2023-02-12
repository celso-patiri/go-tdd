package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestGETPlayers(t *testing.T) {
	is := is.New(t)

	playerStore := StubPlayerStore{
		map[string]int{
			"Celso": 20,
			"Joao":  10,
		},
		[]string{},
		nil,
	}
	server := NewPlayerServer(&playerStore)

	t.Run("get Celso's score", func(t *testing.T) {
		req, res := newRequestAndResponse(http.MethodGet, "/players/Celso")

		server.ServeHTTP(res, req)

		is.Equal(res.Body.String(), "20") // response body is wrong
		is.Equal(res.Code, http.StatusOK) // want status to be 200
	})

	t.Run("get Joao's score", func(t *testing.T) {
		req, res := newRequestAndResponse(http.MethodGet, "/players/Joao")

		server.ServeHTTP(res, req)

		is.Equal(res.Body.String(), "10") // response body is wrong
		is.Equal(res.Code, http.StatusOK) // want status to be 200
	})

	t.Run("return 404 on missing player", func(t *testing.T) {
		req, res := newRequestAndResponse(http.MethodGet, "/players/NIL")

		server.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusNotFound) // want status to be 404
	})

}

func TestScoreWins(t *testing.T) {
	is := is.New(t)

	playerStore := StubPlayerStore{
		map[string]int{},
		[]string{},
		nil,
	}
	server := NewPlayerServer(&playerStore)

	t.Run("it records win on POST", func(t *testing.T) {
		player := "Celso"
		req, res := newRequestAndResponse(http.MethodPost, fmt.Sprintf("/players/%s", player))

		server.ServeHTTP(res, req)

		is.Equal(len(playerStore.winCalls), 1)      // want 1 win to be recorded
		is.Equal(res.Code, http.StatusAccepted)     // want status to be 202
		is.True(playerStore.hasRecordedWin(player)) // want player win to be recorded
	})
}

func TestLeague(t *testing.T) {
	is := is.New(t)

	t.Run("it returns league table as JSON", func(t *testing.T) {
		wantedLeague := []Player{
			{"Celso", 20},
			{"Joao", 10},
		}

		playerStore := StubPlayerStore{nil, nil, wantedLeague}

		server := NewPlayerServer(&playerStore)
		req, res := newRequestAndResponse(http.MethodGet, "/league")

		server.ServeHTTP(res, req)

		got, err := getLeagueFromResponse(t, res.Body)

		is.NoErr(err)                                                      // unable to parse JSON
		is.Equal(res.Result().Header.Get("content-type"), jsonContentType) // should heave application/json header
		is.Equal(wantedLeague, got)                                        // wanted different league JSON object
		is.Equal(res.Code, http.StatusOK)                                  // wanted status code to be 200
	})
}

// Internals

func newRequestAndResponse(method, path string) (req *http.Request, res *httptest.ResponseRecorder) {
	req, _ = http.NewRequest(method, path, nil)
	res = httptest.NewRecorder()
	return req, res
}

func getLeagueFromResponse(t testing.TB, body io.Reader) (league []Player, err error) {
	t.Helper()

	err = json.NewDecoder(body).Decode(&league)

	return league, nil
}

type StubPlayerStore struct {
	scores   map[string]int
	winCalls []string
	league   []Player
}

func (s *StubPlayerStore) GetPlayerScore(name string) (int, bool) {
	score, ok := s.scores[name]
	return score, ok
}

func (s *StubPlayerStore) RecordWin(name string) {
	s.winCalls = append(s.winCalls, name)
}

func (s *StubPlayerStore) hasRecordedWin(name string) bool {
	for _, winner := range s.winCalls {
		if winner == name {
			return true
		}
	}
	return false
}

func (s *StubPlayerStore) GetLeague() []Player {
	return s.league
}

const (
	jsonContentType = "application/json"
)
