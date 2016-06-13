package model

import (
	"appengine"
	"appengine/datastore"
)

// A tag for a game added by a user.
//
// Ancestor: League
type UserGameTag struct {
	User *datastore.Key
	Game *datastore.Key
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

func GetUserGameTags(
	c appengine.Context,
	userKey *datastore.Key,
	leagueKey *datastore.Key,
	gameKey *datastore.Key) ([]*UserGameTag, []*datastore.Key, error) {
	q := datastore.NewQuery("UserGameTag").
		Ancestor(leagueKey)
	if userKey != nil {
		q = q.Filter("User =", userKey)
	}
	if gameKey != nil {
		q = q.Filter("Game =", gameKey)
	}

	var tags []*UserGameTag
	keys, err := q.GetAll(c, &tags)
	if err != nil {
		return nil, nil, err
	}
	return tags, keys, nil
}

func AddUserGameTag(
	c appengine.Context,
	userKey *datastore.Key,
	leagueKey *datastore.Key,
	gameKey *datastore.Key,
	tag string) error {
	userGameTag := &UserGameTag{
		User: userKey,
		Game: gameKey,
		Tag:  tag,
	}
	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		q := datastore.NewQuery("UserGameTag").
			Filter("User =", userKey).
			Filter("Game =", gameKey).
			Filter("Tag =", tag).
			Ancestor(leagueKey).
			Limit(1).
			KeysOnly()
		count, err := q.Count(c)
		if err != nil {
			return err
		}
		if count >= 1 {
			return nil
		}
		key := datastore.NewIncompleteKey(c, "UserGameTag", leagueKey)
		_, err = datastore.Put(c, key, userGameTag)
		return err
	}, nil)
}

func DelUserGameTag(
	c appengine.Context,
	userKey *datastore.Key,
	leagueKey *datastore.Key,
	gameKey *datastore.Key,
	tag string) error {
	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		q := datastore.NewQuery("UserGameTag").
			Filter("User =", userKey).
			Filter("Game =", gameKey).
			Filter("Tag =", tag).
			Ancestor(leagueKey).
			Limit(1).
			KeysOnly()
		keys, err := q.GetAll(c, nil)
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			return datastore.DeleteMulti(c, keys)
		}
		return nil
	}, nil)
}

func GetGameTags(
	c appengine.Context,
	userAcls *RequestorAclCache,
	leagueKey *datastore.Key,
	gameKey *datastore.Key) ([]*GameTag, []*datastore.Key, error) {
	if userAcls != nil {
		if err := userAcls.Can(c, PermissionView, leagueKey); err != nil {
			return nil, nil, err
		}
	}

	q := datastore.NewQuery("GameTag").
		Filter("Game =", gameKey).
		Ancestor(leagueKey)

	var tags []*GameTag
	keys, err := q.GetAll(c, &tags)
	if err != nil {
		return nil, nil, err
	}
	return tags, keys, nil
}

func AddGameTag(
	c appengine.Context,
	userAcls *RequestorAclCache,
	leagueKey *datastore.Key,
	gameKey *datastore.Key,
	tag string,
	reason string) error {
	if userAcls != nil {
		if err := userAcls.Can(c, PermissionEdit, leagueKey); err != nil {
			return err
		}
	}

	gameTag := &GameTag{
		Game:   gameKey,
		Tag:    tag,
		Reason: reason,
	}
	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		q := datastore.NewQuery("GameTag").
			Filter("Game =", gameKey).
			Filter("Tag =", tag).
			Ancestor(leagueKey).
			Limit(1).
			KeysOnly()
		count, err := q.Count(c)
		if err != nil {
			return err
		}
		if count >= 1 {
			return nil
		}
		key := datastore.NewIncompleteKey(c, "GameTag", leagueKey)
		_, err = datastore.Put(c, key, gameTag)
		return err
	}, nil)
}

func DelGameTag(
	c appengine.Context,
	userAcls *RequestorAclCache,
	leagueKey *datastore.Key,
	gameKey *datastore.Key,
	tag string) error {
	if userAcls != nil {
		if err := userAcls.Can(c, PermissionEdit, leagueKey); err != nil {
			return err
		}
	}

	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		q := datastore.NewQuery("GameTag").
			Filter("Game =", gameKey).
			Filter("Tag =", tag).
			Ancestor(leagueKey).
			Limit(1).
			KeysOnly()
		keys, err := q.GetAll(c, nil)
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			return datastore.DeleteMulti(c, keys)
		}
		return nil
	}, nil)
}
