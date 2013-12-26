package model

import (
  "appengine/datastore"
)

type League struct {
  Name string
}

type Team struct {
  Name string
}

type Player struct {
  Summoner string
}

type TeamMembership struct {
  TeamKey *datastore.Key
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
  MatchKey *datastore.Key
  GameResultsKey *datastore.Key
}
