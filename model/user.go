package model

import (
  "appengine"
  "appengine/datastore"
  "appengine/user"
)

const DATASTORE_KIND_USER = "User"

type User struct {
  Name string
  SummonerName string
  Email string
}

// Fetches the user from the datastore if it exists, otherwise puts a new user into
// the datastore and returns it.
func GetUser(c appengine.Context) (*User, error) {
  appengineUser := user.Current(c)

  var user = new(User)
  key := datastore.NewKey(c, DATASTORE_KIND_USER, appengineUser.Email, 0, nil)
  if err := datastore.Get(c, key, user); err == datastore.ErrNoSuchEntity {
    user.Email = appengineUser.Email
    if _, err := datastore.Put(c, key, user); err != nil {
      return nil, err
    }
  } else if err != nil {
    return nil, err
  }
  return user, nil
}
