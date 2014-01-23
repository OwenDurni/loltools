package model

import (
  "appengine"
  "appengine/datastore"
)

type Permission int
const (
  PermissionView = iota
  PermissionEdit
)

// Ancestor: GroupRootKey
type Acl struct {
  // User key or Group key.
  Requestor *datastore.Key
  
  // The key of the protected resource.
  Resource *datastore.Key
  
  Permission Permission
}

func AclAdd(
  c appengine.Context,
  requestor *datastore.Key,
  resource *datastore.Key,
  perm Permission) error {
  groot := GroupRootKey(c)
  
  acl := new(Acl)
  acl.Requestor = requestor
  acl.Resource = resource
  acl.Permission = perm
  
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    q := datastore.NewQuery("Acl").Ancestor(groot).
           Filter("Requestor =", requestor).
           Filter("Resource =", resource).
           Filter("Permission =", perm).
           Limit(1).
           KeysOnly()
    keys, err := q.GetAll(c, nil)
    if err != nil { return err }
    if len(keys) > 0 { return nil }
    key := datastore.NewIncompleteKey(c, "Acl", groot)
    _, err = datastore.Put(c, key, acl)
    return err
  }, nil)
}

func AclRevoke(
  c appengine.Context,
  requestor *datastore.Key,
  resource *datastore.Key,
  perm Permission) error {
  groot := GroupRootKey(c)
  
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    q := datastore.NewQuery("Acl").Ancestor(groot).
           Filter("Requestor =", requestor).
           Filter("Resource =", resource).
           Filter("Permission =", perm).
           Limit(1).
           KeysOnly()
    keys, err := q.GetAll(c, nil)
    if err != nil { return err }
    if len(keys) == 0 { return nil }
    return datastore.Delete(c, keys[0])
  }, nil)
}

type RequestorAclCache struct {
  UserKey *datastore.Key
  EncodedUserKey string
  GroupKeys map[string]*datastore.Key
}
func NewRequestorAclCache(userKey *datastore.Key) *RequestorAclCache {
  r := new(RequestorAclCache)
  r.UserKey = userKey
  r.EncodedUserKey = userKey.Encode()
  return r
}
func (req *RequestorAclCache) init(c appengine.Context) error {
  if req.GroupKeys != nil { return nil }
  memberships, err := GetGroupsForUser(c, req.UserKey)
  if err != nil { return err }
  req.GroupKeys = make(map[string]*datastore.Key)
  for _, m := range memberships {
    req.GroupKeys[m.GroupKey.Encode()] = m.GroupKey
  }
  return err
}

type ResourceAclCache struct {
  ResourceKey *datastore.Key
  EncodedResourceKey string
  Permission Permission
  AuthorizedRequestorKeys map[string]*datastore.Key
}
func NewResourceAclCache(resourceKey *datastore.Key, perm Permission) *ResourceAclCache {
  r := new(ResourceAclCache)
  r.ResourceKey = resourceKey
  r.EncodedResourceKey = resourceKey.Encode()
  r.Permission = perm
  return r
}
func (res *ResourceAclCache) init(c appengine.Context) error {
  if res.AuthorizedRequestorKeys != nil { return nil }
  groot := GroupRootKey(c)
  q := datastore.NewQuery("Acl").Ancestor(groot).
         Filter("Resource =", res.ResourceKey).
         Filter("Permission =", res.Permission)
  var acls []Acl
  _, err := q.GetAll(c, &acls)
  if err != nil { return nil }
  res.AuthorizedRequestorKeys = make(map[string]*datastore.Key)
  for _, acl := range acls {
    res.AuthorizedRequestorKeys[acl.Requestor.Encode()] = acl.Requestor
  }
  return nil
}

func (req *RequestorAclCache) Can(
  perm Permission, res *ResourceAclCache, c appengine.Context) (bool, error) {
  if err := req.init(c); err != nil { return false, err }
  if err := res.init(c); err != nil { return false, err }
  
  // Check if the user is an authorized requestor.
  if _, exists := res.AuthorizedRequestorKeys[req.EncodedUserKey]; exists {
    return true, nil
  }
  
  // Check if the user is in a group that is an authorized requestor.
  for encodedRequestor := range req.GroupKeys {
    if _, exists := res.AuthorizedRequestorKeys[encodedRequestor]; exists {
      return true, nil
    }
  }
  
  // User is not authorized.
  return false, nil
}