package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
)

// Leagues are identified by their key.
type League struct {
  // The name of the league. Not necessarily unique.
  Name string

  // The datastore key for the User who owns this league.
  Owner *datastore.Key
}

type Team struct {
  Name string
}

type Player struct {
  Summoner string
}

type TeamMembership struct {
  TeamKey   *datastore.Key
  PlayerKey *datastore.Key
}

type Match struct {
  // Short string describing the purpose of the match. (Ex: "Week 1", "Round of 64")
  Tag string
  // The match consists of this number of games. (Ex: 3 for a best-of-3).
  GameCount int

  // The two teams involved in the match.
  Team1 *datastore.Key
  Team2 *datastore.Key
}

type MatchResults struct {
  MatchKey       *datastore.Key
  GameResultsKey *datastore.Key
}

func CreateLeague(c appengine.Context, name string) (*League, *datastore.Key, error) {
  _, userKey, err := GetUser(c)
  if err != nil {
    return nil, nil, err
  }

  var league = new(League)
  league.Name = name
  league.Owner = userKey
  leagueKey := datastore.NewIncompleteKey(c, "League", nil)

  leagueKey, err = datastore.Put(c, leagueKey, league)
  if err != nil {
    return nil, nil, err
  }
  return league, leagueKey, nil
}

func LeagueUri(leagueKey *datastore.Key) string {
  return fmt.Sprintf("/leagues/%v", EncodeGlobalKeyShort(leagueKey))
}

type LeagueInfo struct {
  League League
  LeagueKey *datastore.Key
}

func LeaguesForUser(
    c appengine.Context, userKey *datastore.Key) (result []*LeagueInfo, err error) {
  result = nil
  err = nil
  if userKey == nil {
    return
  }
  
  q := datastore.NewQuery("League").
           Filter("Owner =", userKey).
           Order("Name")
  var leagues []League
  leagueKeys, err := q.GetAll(c, &leagues)
  if err != nil {
    return
  }
  result = make([]*LeagueInfo, len(leagues))
  for i := range leagues {
    info := new(LeagueInfo)
    info.League = leagues[i]
    info.LeagueKey = leagueKeys[i]
    result[i] = info
  }
  return
}