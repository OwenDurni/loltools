package model

import (
  "appengine/datastore"
  "time"
)

// A scheduled match between two teams. May consist of one or more ScheduledGames
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
  InitialDatetime *time.Time
  
  // The earliest the match should be played. This is not enforced.
  DateEarliest *time.Time
  
  // The latest the match should be played. This is not enforced.
  DateLatest   *time.Time
}