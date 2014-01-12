package model

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/util/errwrap"
)

// Leagues are identified by their key.
type League struct {
  // The name of the league. Not necessarily unique.
  Name string

  // The Riot server region for this league. Defaults to RegionNA.
  Region string

  // The datastore key for the User who owns this league.
  Owner *datastore.Key
}

// Teams are identified by their datastore.Key.
//
// Ancestor: League
type Team struct {
  Name string
}

// A table associating summoners to teams. Summoners may be on more than one team
// per league. Teams may have any number of summoners.
//
// Ancestor: League
type TeamMembership struct {
  TeamKey   *datastore.Key
  PlayerKey *datastore.Key
}

// Various accumulated data about a team. Not directly stored in datastore.
type TeamInfo struct {
  Name string
  Id   string
  Uri  string
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
    return nil, nil, errwrap.Wrap(err)
  }

  var league = new(League)
  league.Name = name
  league.Owner = userKey
  league.Region = RegionNA
  leagueKey := datastore.NewIncompleteKey(c, "League", nil)

  leagueKey, err = datastore.Put(c, leagueKey, league)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }
  return league, leagueKey, nil
}

func LeagueUri(leagueKey *datastore.Key) string {
  return fmt.Sprintf("/leagues/%s", EncodeKeyShort(leagueKey))
}

func LeagueTeamUri(leagueKey *datastore.Key, teamKey *datastore.Key) string {
  return fmt.Sprintf("%s/teams/%s",
    LeagueUri(leagueKey),
    EncodeKeyShort(teamKey))
}

type LeagueInfo struct {
  League    League
  LeagueKey *datastore.Key
}

func LeaguesForUser(
  c appengine.Context, userKey *datastore.Key) ([]*LeagueInfo, error) {
  q := datastore.NewQuery("League").
    Filter("Owner =", userKey).
    Order("Name")
  var leagues []League
  leagueKeys, err := q.GetAll(c, &leagues)
  if err != nil {
    return nil, errwrap.Wrap(err)
  }
  result := make([]*LeagueInfo, len(leagues))
  for i := range leagues {
    info := new(LeagueInfo)
    info.League = leagues[i]
    info.LeagueKey = leagueKeys[i]
    result[i] = info
  }
  return result, nil
}

func LeagueById(
  c appengine.Context,
  userKey *datastore.Key,
  leagueId string) (*League, *datastore.Key, error) {
  leagueKey, err := DecodeKeyShort(c, "League", leagueId, nil)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  league := new(League)
  if err := datastore.Get(c, leagueKey, league); err != nil {
    return nil, leagueKey, errwrap.Wrap(err)
  }

  // TODO(durni): Check viewing permissions
  return league, leagueKey, nil
}

func TeamById(
  c appengine.Context,
  userKey *datastore.Key,
  leagueKey *datastore.Key,
  teamId string) (*Team, *datastore.Key, error) {
  teamKey, err := DecodeKeyShort(c, "Team", teamId, leagueKey)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  team := new(Team)
  if err := datastore.Get(c, teamKey, team); err != nil {
    return nil, teamKey, errwrap.Wrap(err)
  }

  // TODO(durni): Check viewing permissions
  return team, teamKey, nil
}

func LeagueAddTeam(
  c appengine.Context,
  userKey *datastore.Key,
  leagueId string,
  teamName string) (*Team, *datastore.Key, error) {
  // TODO(durni): Check that this user has permissions to add teams to the league.

  _, leagueKey, err := LeagueById(c, userKey, leagueId)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  team := new(Team)
  team.Name = teamName
  var teamKey *datastore.Key

  err = datastore.RunInTransaction(c, func(c appengine.Context) error {
    q := datastore.NewQuery("Team").Ancestor(leagueKey).
      Filter("Name =", team.Name)
    var teams []Team
    if _, err := q.GetAll(c, &teams); err != nil {
      return err
    }
    if len(teams) > 0 {
      return errors.New(fmt.Sprintf("team already exists: %v", teams[0].Name))
    }
    teamKey = datastore.NewIncompleteKey(c, "Team", leagueKey)
    teamKey, err = datastore.Put(c, teamKey, team)
    if err != nil {
      return errwrap.Wrap(err)
    }
    return nil
  }, nil)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }
  return team, teamKey, nil
}

func LeagueAllTeams(
  c appengine.Context,
  userKey *datastore.Key,
  leagueKey *datastore.Key) ([]*Team, []*datastore.Key, error) {
  var teams []*Team
  var teamKeys []*datastore.Key
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    q := datastore.NewQuery("Team").Ancestor(leagueKey)
    var err error
    teamKeys, err = q.GetAll(c, &teams)
    return errwrap.Wrap(err)
  }, nil)
  return teams, teamKeys, errwrap.Wrap(err)
}

func TeamAllPlayers(
  c appengine.Context,
  userKey *datastore.Key,
  leagueKey *datastore.Key,
  teamKey *datastore.Key,
  keysOnly KeysOnlyOption) ([]*Player, []*datastore.Key, error) {
  var memberships []TeamMembership
  q := datastore.NewQuery("TeamMembership").Ancestor(leagueKey).
    Filter("TeamKey =", teamKey)
  _, err := q.GetAll(c, &memberships)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  playerKeys := make([]*datastore.Key, len(memberships))
  for i, m := range memberships {
    playerKeys[i] = m.PlayerKey
  }
  if keysOnly == KeysOnly {
    return nil, playerKeys, nil
  }

  players := make([]*Player, len(playerKeys))
  for i := range players {
    players[i] = new(Player)
  }
  err = datastore.GetMulti(c, playerKeys, players)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }
  return players, playerKeys, nil
}

func TeamAddPlayer(
  c appengine.Context,
  userKey *datastore.Key,
  leagueKey *datastore.Key,
  teamKey *datastore.Key,
  playerKey *datastore.Key) error {
  m := &TeamMembership{
    TeamKey: teamKey,
    PlayerKey: playerKey,
  }
  key := datastore.NewIncompleteKey(c, "TeamMembership", leagueKey)
  _, err := datastore.Put(c, key, m)
  return err
}
