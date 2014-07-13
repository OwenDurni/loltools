package model

import (
  "appengine"
  "appengine/datastore"
  "appengine/user"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
  "github.com/OwenDurni/loltools/util/errwrap"
  "math/rand"
  "time"
)

type User struct {
  Email string
  
  // Empty if no verified summoner.
  DisplayName string
}

// Key: ("%s:%s", User.StringID(), Player.StringID())
type UnverifiedSummoner struct {
  User       *datastore.Key
  Player     *datastore.Key
  Token      string
  CreateTime time.Time
}

// Key: ("%s:%s", User.StringID(), Player.StringID())
type VerifiedSummoner struct {
  User   *datastore.Key
  Player *datastore.Key
}

type SummonerData struct {
  Player   *Player
  Verified bool
  Token    string
}

// Fetches the user from the datastore if it exists, otherwise puts a new user into
// the datastore and returns it.
func GetUser(c appengine.Context) (*User, *datastore.Key, error) {
  appengineUser := user.Current(c)

  user := new(User)
  key := datastore.NewKey(c, "User", appengineUser.Email, 0, nil)
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    err := datastore.Get(c, key, user)
    if err == datastore.ErrNoSuchEntity {
      user.Email = appengineUser.Email
      _, err = datastore.Put(c, key, user)
    }
    return err
  }, nil)
  return user, key, err
}

func GetUserByKey(c appengine.Context, userKey *datastore.Key) (*User, error) {
  var user = new(User)
  if err := datastore.Get(c, userKey, user); err != nil {
    return nil, err
  }
  return user, nil
}

func GetUserByEmail(c appengine.Context, email string) (*User, *datastore.Key, error) {
  q := datastore.NewQuery("User").
    Filter("Email =", email).
    Limit(1)
  var users []*User
  userKeys, err := q.GetAll(c, &users)
  if err != nil {
    return nil, nil, err
  }
  if len(userKeys) == 0 {
    return nil, nil, errors.New(fmt.Sprintf("User does not exist: %s", email))
  }
  return users[0], userKeys[0], nil
}

func GetSummonerDatas(c appengine.Context, userKey *datastore.Key) ([]*SummonerData, error) {
  // Get unverified summoners for this user.
  q := datastore.NewQuery("UnverifiedSummoner").
         Filter("User =", userKey)
  var unverifiedSummoners []UnverifiedSummoner
  _, err := q.GetAll(c, &unverifiedSummoners)
  if err != nil {
    return nil, errwrap.Wrap(err)
  }
  
  // Get verified summoners for this user.
  q = datastore.NewQuery("VerifiedSummoner").
        Filter("User =", userKey)
  var verifiedSummoners []VerifiedSummoner
  _, err = q.GetAll(c, &verifiedSummoners)
  if err != nil {
    return nil, errwrap.Wrap(err)
  }
  
  // Collect all player keys so we can look up corresponding players.
  numPlayers := len(unverifiedSummoners)+len(verifiedSummoners)
  ret := make([]*SummonerData, numPlayers)
  playerKeys := make([]*datastore.Key, numPlayers)
  players := make([]*Player, numPlayers)
  
  r := 0
  for _, s := range verifiedSummoners {
    ret[r] = &SummonerData{Player: nil, Verified: true, Token: ""}
    players[r] = new(Player)
    playerKeys[r] = s.Player
    r++
  }
  for _, s := range unverifiedSummoners {
    ret[r] = &SummonerData{Player: nil, Verified: false, Token: s.Token}
    players[r] = new(Player)
    playerKeys[r] = s.Player
    r++
  }
  
  // Lookup players.
  err = datastore.GetMulti(c, playerKeys, players)
  if err != nil {
    return nil, errwrap.Wrap(err)
  }
  
  // Populate players
  for r := 0; r < numPlayers; r++ {
    ret[r].Player = players[r]
  }
  
  return ret, nil
}

func VerifySummoner(
  c appengine.Context,
  userKey *datastore.Key,
  playerKey *datastore.Key,
  player *Player) error {
  // Get the unverified summoner if one exists.
  keyName := fmt.Sprintf("%s:%s", userKey.StringID(), playerKey.StringID())
  unverifiedKey := datastore.NewKey(c, "UnverifiedSummoner", keyName, 0, nil)
  unverifiedSummoner := new(UnverifiedSummoner)
  err := datastore.Get(c, unverifiedKey, unverifiedSummoner)
  if err != nil {
    return err
  }
  
  // Lookup rune pages for player.
  riotApiKey, err := GetRiotApiKey(c)
  if err != nil {
    return err
  }
  if err = RiotApiRateLimiter.Consume(c, 1); err != nil {
    return err
  }
  runePagesDto, err := riot.RunesBySummonerId(c, riotApiKey.Key, player.Region, player.RiotId)
  if err != nil {
    return err
  }
  
  // Find rune page with name matching code to verify.
  for _, runePageDto := range runePagesDto.Pages {
    if runePageDto.Name == unverifiedSummoner.Token {
      verifiedSummonerKey := datastore.NewKey(c, "VerifiedSummoner", keyName, 0, nil)
      verifiedSummoner := new(VerifiedSummoner)
      verifiedSummoner.User = userKey
      verifiedSummoner.Player = playerKey
      
      // Remove the unverified summoner and add a verified one.
      err = datastore.RunInTransaction(c, func(c appengine.Context) error {
        err = datastore.Get(c, unverifiedKey, unverifiedSummoner)
        if err != nil {
          return err
        }
        err = datastore.Delete(c, unverifiedKey)
        if err != nil {
          return err
        }
        _, err := datastore.Put(c, verifiedSummonerKey, verifiedSummoner)
        return err
      }, &datastore.TransactionOptions{XG: true})
      if err != nil {
        return err
      }
      return err
    }
  }
    
  return errors.New(fmt.Sprintf("error: %s did not have a runepage with the correct code", player.Summoner))
}

func AddUnverifiedSummoner(
  c appengine.Context,
  userKey *datastore.Key,
  region string,
  summoner string) error {
  _, playerKey, err := GetOrCreatePlayerBySummoner(c, region, summoner)
  if err != nil {
    return err
  }
  
  keyName := fmt.Sprintf("%s:%s", userKey.StringID(), playerKey.StringID())
  unverifiedKey := datastore.NewKey(c, "UnverifiedSummoner", keyName, 0, nil)
  verifiedKey := datastore.NewKey(c, "VerifiedSummoner", keyName, 0, nil)
  unverifiedSummoner := new(UnverifiedSummoner)
  verifiedSummoner := new(VerifiedSummoner)
  
  now := time.Now()
  rnd := rand.New(rand.NewSource(now.Unix()))
  token := ""
  for i := 0; i < 9; i++ {
    token += fmt.Sprintf("%d", rnd.Intn(10))
  }
  
  // Add an unverified summoner only if neither a verified summoner nor an
  // unverified summoner exist.
  err = datastore.RunInTransaction(c, func(c appengine.Context) error {
    err := datastore.Get(c, verifiedKey, verifiedSummoner)
    if err == nil {
      return nil
    } else if err != datastore.ErrNoSuchEntity {
      return err
    }
    
    err = datastore.Get(c, unverifiedKey, unverifiedSummoner)
    if err == nil {
      return nil
    } else if err != datastore.ErrNoSuchEntity {
      return err
    }
    
    unverifiedSummoner.User = userKey
    unverifiedSummoner.Player = playerKey
    unverifiedSummoner.Token = token
    unverifiedSummoner.CreateTime = now
    
    _, err = datastore.Put(c, unverifiedKey, unverifiedSummoner)
    return err
  }, &datastore.TransactionOptions{XG: true})
  return err
}

func SetPrimarySummoner(
  c appengine.Context,
  userKey *datastore.Key,
  player *Player,
  playerKey *datastore.Key) error {
  // Ensure the specified player is a verified summoner for this user.
  q := datastore.NewQuery("VerifiedSummoner").
      Filter("User =", userKey).
      Filter("Player =", playerKey).
      Limit(1).
      KeysOnly()
  count, err := q.Count(c)
  if err != nil {
    return err
  }
  if count == 0 {
    return errors.New(fmt.Sprintf("You must verify summoner %s first", player.Summoner))
  }
    
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    user := new(User)
    err := datastore.Get(c, userKey, user)
    if err != nil {
      return err
    }
    user.DisplayName = fmt.Sprintf("%s-%s", player.Region, player.Summoner)
    _, err = datastore.Put(c, userKey, user)
    return err
  }, nil)
}
  
