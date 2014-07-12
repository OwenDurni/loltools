package model

import (
  "appengine"
  "appengine/datastore"
)

// A tag for a game added by a user.
//
// Ancestor: League
type UserGameTag struct {
  Game *datastore.Key
  User *datastore.Key
  Tag  string
}

// A tag for a game added by the system.
//
// Ancestor: League
type GameTag struct {
  Game   *datastore.Key
  Tag    string
  Reason string
}

func AddUserGameTag(
  c appengine.Context,
  userKey *datastore.Key,
  leagueKey *datastore.Key,
  gameKey *datastore.Key,
  tag string) error {
  return nil
}

func DelUserGameTag(
  c appengine.Context,
  userKey *datastore.Key,
  leagueKey *datastore.Key,
  gameKey *datastore.Key,
  tag string) error {
  return nil
}

func AddGameTag(
  c appengine.Context,
  userAcls *RequestorAclCache,
  leagueKey *datastore.Key,
  gameKey *datastore.Key,
  tag string,
  reason string) error {
  return nil
}

func DelGameTag(
  c appengine.Context,
  userAcls *RequestorAclCache,
  leagueKey *datastore.Key,
  gameKey *datastore.Key,
  tag string) error {
  return nil
}