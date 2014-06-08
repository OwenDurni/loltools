package model

type Score struct {
  // The number of points earned.
  Points  int
  // True if the score will not change if additional games are played.
  IsFinal bool
}

// A TeamScoring is used to take a chronological series of games and compute how many points
// a team has earned for those results. Results may be negative.
type TeamScoring interface {
  Score(match *ScheduledMatch, games []*GameInfo) Score
}

// A PlayerScoring is used to take a chronological series of games and compute how many points
// a player has earned for those results. Results may be negative.
type PlayerScoring interface {
  Score(match *ScheduledMatch, player *Player, games []*GameInfo) Score
}

// A list of supported scoring algorithms for matches.
//
// In the descriptions below, let
//   M := The number of games in the match (or if the length is unbounded, the number of games played),
//   G := The list of games played,
//   S := The first M games of G,
//   T := The team of interest.
type TeamScoringType int
const (
  // BestOfSeries
  // ------------
  // Points
  //   2 points: T wins a majority of games in S
  //   1 point : T wins exactly M/2 of games in S (if M is even)
  //   0 points: otherwise
  //
  // IsFinal
  //   true : at least M games have been played || T wins or loses a majority of games in S
  //   false: otherwise
  BestOfSeries TeamScoringType = 0
  
  // ThreeOneZeroSeries
  // ------------------
  // Points
  // 3 points: T wins a majority of games in S
  // 1 point : T wins exactly M/2 of games in S (if M is even)
  // 0 points: otherwise
  //
  // IsFinal
  // true : at least M games have been played
  // false: otherwise
  ThreeOneZeroSeries TeamScoringType = 1
)