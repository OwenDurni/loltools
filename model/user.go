package model

import (
  "appengine"
  "appengine/datastore"
  "appengine/user"
  "errors"
  "fmt"
  "math/rand"
  "time"
)

type User struct {
  Email string
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
