package model

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/util/errwrap"
  "time"
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

// An association between games and teams.
// 
// Specifically the games containing at least 3 members of the league-team on the
// same in-game team.
//
// Ancestor: League
type GameByTeam struct {
  GameKey  *datastore.Key
  TeamKey  *datastore.Key
  DateTime time.Time
}

// An associate between games and tags. A game can have multiple tags.
//
// Ancestor: League
type GameTags struct {
  GameKey *datastore.Key
  Tag     string
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

func LeaguesForUser(
  c appengine.Context, userAcls *RequestorAclCache) ([]*League, []*datastore.Key, error) {
  retLeagues := make([]*League, 0, 8)
  retLeagueKeys := make([]*datastore.Key, 0, 8)
  
  // Leagues owned.
  {
    q := datastore.NewQuery("League").
      Filter("Owner =", userAcls.UserKey).
      Order("Name")
    var leagues []*League
    leagueKeys, err := q.GetAll(c, &leagues)
    if err != nil {
      return nil, nil, err
    }
    retLeagues = append(retLeagues, leagues...)
    retLeagueKeys = append(retLeagueKeys, leagueKeys...)
  }
  
  leagueKeys, err := userAcls.FindAll(c, "League", PermissionView)
  if err != nil { return nil, nil, err }
    
  leagues := make([]*League, len(leagueKeys))
  for i := range leagues {
    leagues[i] = new(League)
  }
  err = datastore.GetMulti(c, leagueKeys, leagues)
  if err != nil { return nil, nil, err }
        
  retLeagues = append(retLeagues, leagues...)
  retLeagueKeys = append(retLeagueKeys, leagueKeys...)
  
  return retLeagues, retLeagueKeys, nil
}

func LeagueById(
  c appengine.Context,
  leagueId string) (*League, *datastore.Key, error) {
  leagueKey, err := DecodeKeyShort(c, "League", leagueId, nil)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  league := new(League)
  if err := datastore.Get(c, leagueKey, league); err != nil {
    return nil, leagueKey, errwrap.Wrap(err)
  }

  return league, leagueKey, nil
}

func TeamById(
  c appengine.Context,
  userAcls *RequestorAclCache,
  league *League,
  leagueKey *datastore.Key,
  teamId string) (*Team, *datastore.Key, error) {
  teamKey, err := DecodeKeyShort(c, "Team", teamId, leagueKey)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  if userAcls != nil {
    if *userAcls.UserKey != *league.Owner {
      if err = userAcls.Can(c, PermissionView, leagueKey); err != nil {
        return nil, nil, err
      }
    }
  }
  
  team := new(Team)
  if err := datastore.Get(c, teamKey, team); err != nil {
    return nil, teamKey, errwrap.Wrap(err)
  }

  return team, teamKey, nil
}

func LeagueAddTeam(
  c appengine.Context,
  userAcls *RequestorAclCache,
  leagueId string,
  teamName string) (*Team, *datastore.Key, error) {
  league, leagueKey, err := LeagueById(c, leagueId)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  if userAcls != nil {
    if *userAcls.UserKey != *league.Owner {
      if err = userAcls.Can(c, PermissionEdit, leagueKey); err != nil {
        return nil, nil, err
      }
    }
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
  userAcls *RequestorAclCache,
  league *League,
  leagueKey *datastore.Key) ([]*Team, []*datastore.Key, error) {
  
  if userAcls != nil {
    if *userAcls.UserKey != *league.Owner {
      if err := userAcls.Can(c, PermissionView, leagueKey); err != nil {
        return nil, nil, err
      }
    }
  }
  
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
  userAcls *RequestorAclCache,
  league *League,
  leagueKey *datastore.Key,
  teamKey *datastore.Key,
  keysOnly KeysOnlyOption) ([]*Player, []*datastore.Key, error) {
  
  if userAcls != nil {
    if *userAcls.UserKey != *league.Owner {
      if err := userAcls.Can(c, PermissionView, leagueKey); err != nil {
        return nil, nil, err
      }
    }
  }
    
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
  userAcls *RequestorAclCache,
  league *League,
  leagueKey *datastore.Key,
  teamKey *datastore.Key,
  playerKey *datastore.Key) error {
    
  if userAcls != nil {
    if *userAcls.UserKey != *league.Owner {
      if err := userAcls.Can(c, PermissionEdit, leagueKey); err != nil {
        return err
      }
    }
  }
    
  m := &TeamMembership{
    TeamKey:   teamKey,
    PlayerKey: playerKey,
  }
  key := datastore.NewIncompleteKey(c, "TeamMembership", leagueKey)
  _, err := datastore.Put(c, key, m)
  return err
}

func TeamDelPlayer(
  c appengine.Context,
  userAcls *RequestorAclCache,
  league *League,
  leagueKey *datastore.Key,
  teamKey *datastore.Key,
  playerKey *datastore.Key) error {
    
  if userAcls != nil {
    if *userAcls.UserKey != *league.Owner {
      if err := userAcls.Can(c, PermissionEdit, leagueKey); err != nil {
        return err
      }
    }
  }
    
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    q := datastore.NewQuery("TeamMembership").Ancestor(leagueKey).
           Filter("TeamKey =", teamKey).
           Filter("PlayerKey =", playerKey).
           Limit(1).
           KeysOnly()
    keys, err := q.GetAll(c, nil)
    if err != nil { return err }
    if len(keys) > 0 {
      return datastore.Delete(c, keys[0])
    }
    return nil
  }, nil)
}

func LeagueAddGameByTeam(
  c appengine.Context,
  leagueKey *datastore.Key,
  gameKey *datastore.Key,
  teamKey *datastore.Key,
  dateTime time.Time) error {
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    q := datastore.NewQuery("GameByTeam").Ancestor(leagueKey).
           Filter("GameKey =", gameKey).
           Filter("TeamKey =", teamKey).
           Limit(1).
           KeysOnly()
    keys, err := q.GetAll(c, nil)
    if err != nil { return err }
    if len(keys) > 0 { return nil }
    key := datastore.NewIncompleteKey(c, "GameByTeam", leagueKey)
    gameByTeam := new(GameByTeam)
    gameByTeam.GameKey = gameKey
    gameByTeam.TeamKey = teamKey
    gameByTeam.DateTime = dateTime
    _, err = datastore.Put(c, key, gameByTeam)
    return err
  }, nil)
  return err
}