package model

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
  "time"
)

// A scheduled match between two teams.
//
// Ancestor: League
type ScheduledMatch struct {
  // A brief "one-line description" of the match.
  // Example: "Summoner's Rift Draft (Best of 1)"
  Summary string
  
  // A more detailed description of the match.
  Description string
  
  // The primary tag used to group this with other matches.
  // Example: "Week 1"
  PrimaryTag string
  
  // The teams involved in the match.
  TeamKeys []*datastore.Key
  
  // The number of games in the match. Zero or less means "no limit".
  NumGames int
  
  // The official suggested date(time) of the match.
  OfficialDatetime time.Time
  
  // The earliest the match should be played. This is not enforced.
  DateEarliest time.Time
  
  // The latest the match should be played. This is not enforced.
  DateLatest   time.Time
}
func (m *ScheduledMatch) HomeTeam() *datastore.Key {
  return m.TeamKeys[0]
}
func (m *ScheduledMatch) AwayTeam() *datastore.Key {
  return m.TeamKeys[1]
}
func MatchId(matchKey *datastore.Key) string {
  return EncodeKeyShort(matchKey)
}

type MatchResult struct {
  ScheduledMatch *datastore.Key
  Team           *datastore.Key
  Points         int
  ManualResult   bool
}

// Creates a scheduled match.
func CreateScheduledMatch(
  c appengine.Context,
  userAcls *RequestorAclCache,
  league *League,
  leagueKey *datastore.Key,
  match *ScheduledMatch) error {
  c.Debugf("model.CreateScheduledMatch begin")
  // Creating matches requires edit permissions on the league.
  if userAcls != nil {
    if *userAcls.UserKey != *league.Owner {
      if err := userAcls.Can(c, PermissionEdit, leagueKey); err != nil {
        return err
      }
    }
  }
  
  // Ensure specified teams are in this league.
  for _, teamKey := range match.TeamKeys {
    if *teamKey.Parent() != *leagueKey {
      return errors.New(fmt.Sprintf(
        "team '%s' not in league '%s'", teamKey.String(), leagueKey.String()))
    }
  }
  
  _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "ScheduledMatch", leagueKey), match)
  c.Debugf("model.CreateScheduledMatch end")
  return err
}