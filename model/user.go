package model

import (
  "appengine"
  "appengine/datastore"
  "appengine/user"
  "errors"
)

type User struct {
  Name         string
  SummonerName string
  Email        string
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

func (user *User) Save(c appengine.Context) error {
  if user == nil {
    return errors.New("nil user")
  }
  key := datastore.NewKey(c, "User", user.Email, 0, nil)
  _, err := datastore.Put(c, key, user)
  return err
}
