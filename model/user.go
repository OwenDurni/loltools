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

  var user = new(User)
  key := datastore.NewKey(c, "User", appengineUser.Email, 0, nil)
  if err := datastore.Get(c, key, user); err == datastore.ErrNoSuchEntity {
    user.Email = appengineUser.Email
    if _, err := datastore.Put(c, key, user); err != nil {
      return nil, key, err
    }
  } else if err != nil {
    return nil, key, err
  }
  return user, key, nil
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
