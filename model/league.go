package model

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
)

// Leagues are identified by their key.
type League struct {
  // The name of the league. Not necessarily unique.
  Name string

  // The datastore key for the User who owns this league.
  Owner *datastore.Key
}

// Teams are identified by their datastore.Key and are a child of the League.
type Team struct {
  Name string
}

// Various accumulated data about a team. Not directly stored in datastore.
type TeamInfo struct {
  Name string
  Id string
  Uri string 
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
  return fmt.Sprintf("/leagues/%s", EncodeKeyShort(leagueKey))
}

func LeagueTeamUri(leagueKey *datastore.Key, teamKey *datastore.Key) string {
  return fmt.Sprintf("%s/teams/%s",
                     LeagueUri(leagueKey),
                     EncodeKeyShort(teamKey))
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

func LeagueById(
    c appengine.Context,
    userKey *datastore.Key,
    leagueId string) (*League, *datastore.Key, error) {
  if userKey == nil {
    return nil, nil, errors.New("LeagueById(): nil userKey")
  }
  
  leagueKey, err := DecodeKeyShort(c, "League", leagueId, nil)
  if err != nil {
    return nil, nil, err
  }
  
  league := new(League)
  if err := datastore.Get(c, leagueKey, league); err != nil {
    return nil, leagueKey, err
  }
  
  // TODO(durni): Check viewing permissions
  return league, leagueKey, nil
}

func LeagueAddTeam(
    c appengine.Context,
    userKey *datastore.Key,
    leagueId string,
    teamName string) (*Team, *datastore.Key, error) {
  // TODO(durni): Check that this user has permissions to add teams to the league.
  
  _, leagueKey, err := LeagueById(c, userKey, leagueId)
  if err != nil {
    return nil, nil, err
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
      return err
    }
    return nil
  }, nil)
  if err != nil {
    return nil, nil, err
  }
  return team, teamKey, nil
}

func LeagueAllTeams(
    c appengine.Context,
    userKey *datastore.Key,
    leagueKey *datastore.Key) ([]*TeamInfo, error) {
  var infos []*TeamInfo
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    var teams []Team
    q := datastore.NewQuery("Team").Ancestor(leagueKey)
    teamKeys, err := q.GetAll(c, &teams)
    if err != nil {
      return err
    }
    infos = make([]*TeamInfo, len(teams))
    for i, team := range teams {
      t := new(TeamInfo)
      t.Name = team.Name
      t.Id = EncodeKeyShort(teamKeys[i])
      t.Uri = LeagueTeamUri(leagueKey, teamKeys[i])      
      infos[i] = t
    }
    return nil
  }, nil)
  return infos, err
}