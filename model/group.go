package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
)

// Root entity for all groups and acls.
type GroupRoot struct {}
func GroupRootKey(c appengine.Context) *datastore.Key {
  return datastore.NewKey(c, "GroupRoot", "dev", 0, nil)
}

type Group struct {
  Name    string
}

type GroupMembership struct {
  GroupKey *datastore.Key
  UserKey  *datastore.Key
  Owner    bool
}

func GroupId(groupKey *datastore.Key) string {
  return EncodeKeyShort(groupKey)
}

func GroupUri(groupKey *datastore.Key) string {
  return fmt.Sprintf("/groups/%s", GroupId(groupKey))
}

func GetGroupByKey(c appengine.Context, groupKey *datastore.Key) (*Group, error) {
  g := new(Group)
  err := datastore.Get(c, groupKey, g)
  return g, err
}

func GetGroupsForUser(c appengine.Context, userKey *datastore.Key) ([]*GroupMembership, error) {
  q := datastore.NewQuery("GroupMembership").Ancestor(GroupRootKey(c)).
    Filter("UserKey =", userKey)
  
  var memberships []*GroupMembership
  _, err := q.GetAll(c, &memberships)
  return memberships, err
}

// Creates a new group with the current user as the owner.
func CreateGroup(c appengine.Context, name string) (*Group, *datastore.Key, error) {
  _, userKey, err := GetUser(c)
  if err != nil {
    return nil, nil, err
  }

  groot := GroupRootKey(c)
  
  group := new(Group)
  group.Name = name
  var groupKey *datastore.Key
  
  err = datastore.RunInTransaction(c, func(c appengine.Context) error {
    groupKey, err = datastore.Put(c, datastore.NewIncompleteKey(c, "Group", groot), group)
    if err != nil {
      return err
    }

    groupMembership := &GroupMembership{
      GroupKey: groupKey,
      UserKey:  userKey,
      Owner:    true,
    }
    _, err = datastore.Put(c, datastore.NewIncompleteKey(c, "GroupMembership", groot),
                           groupMembership)
    return err
  }, nil)
  if err != nil {
    return nil, nil, err
  }
  return group, groupKey, nil
}
