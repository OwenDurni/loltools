package model

import (
  "appengine"
  "appengine/datastore"
  "appengine/user"
  "fmt"
)

type ErrNotAuthorized struct {
  Permission Permission
  Resource   *datastore.Key
}
func (e ErrNotAuthorized) Error() string {
  return fmt.Sprintf("You are not authorized to %s this %s",
                     e.Permission.String(), e.Resource.Kind())
}

type Permission int
const (
  PermissionView = iota
  PermissionEdit
)
func (p Permission) String() string {
  if p == PermissionView { return "view" }
  if p == PermissionEdit { return "edit" }
  return "<unknown_operation>"
}
func AllPermissions() []Permission {
  return []Permission{PermissionView, PermissionEdit}
}

// Ancestor: GroupRootKey
type Acl struct {
  // User key or Group key.
  Requestor *datastore.Key
  
  // The key of the protected resource.
  Resource *datastore.Key
  
  // The entity type of the resource key.
  ResourceKind string
  
  Permission Permission
}

func AclCan(
  c appengine.Context,
  requestor *datastore.Key,
  perm Permission,
  resource *datastore.Key) (bool, error) {
  groot := GroupRootKey(c)
  q := datastore.NewQuery("Acl").Ancestor(groot).
         Filter("Requestor =", requestor).
         Filter("Resource =", resource).
         Filter("Permission =", perm).
         Limit(1).
         KeysOnly()
  keys, err := q.GetAll(c, nil)
  return len(keys) > 0, err
}

func AclFindAll(
  c appengine.Context,
  requestor *datastore.Key,
  resourceKind string,
  perm Permission) ([]*datastore.Key, error) {
  groot := GroupRootKey(c)
  q := datastore.NewQuery("Acl").Ancestor(groot).
         Filter("Requestor =", requestor).
         Filter("ResourceKind =", resourceKind).
         Filter("Permission =", perm).
         Project("Resource")
  var acls []*Acl
  if _, err := q.GetAll(c, &acls); err != nil {
    return nil, err
  }
  resources := make([]*datastore.Key, len(acls))
  for i := range acls {
    resources[i] = acls[i].Resource
  }
  return resources, nil
}

func AclGrant(
  c appengine.Context,
  requestor *datastore.Key,
  resource *datastore.Key,
  perm Permission) error {
  groot := GroupRootKey(c)
  
  acl := new(Acl)
  acl.Requestor = requestor
  acl.Resource = resource
  acl.ResourceKind = resource.Kind()
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
  Groups map[string]*Group
  GroupKeys map[string]*datastore.Key
  
  resCache map[Permission]map[string]*ResourceAclCache
}
func NewRequestorAclCache(userKey *datastore.Key) *RequestorAclCache {
  r := new(RequestorAclCache)
  r.UserKey = userKey
  r.EncodedUserKey = userKey.Encode()
  return r
}
func (req *RequestorAclCache) init(c appengine.Context) error {
  if req.GroupKeys != nil { return nil }
  groups, memberships, err := GetGroupsForUser(c, req.UserKey)
  if err != nil { return err }
  req.Groups = make(map[string]*Group)
  req.GroupKeys = make(map[string]*datastore.Key)
  for i, m := range memberships {
    req.GroupKeys[m.GroupKey.Encode()] = m.GroupKey
    req.Groups[m.GroupKey.Encode()] = groups[i]
  }
  return err
}
func (req *RequestorAclCache) lookupResourceAcls(
  perm Permission, resKey *datastore.Key) *ResourceAclCache {
  encodedResKey := resKey.Encode()
    
  if req.resCache == nil {
    req.resCache = make(map[Permission]map[string]*ResourceAclCache)
  }
  caches, exists := req.resCache[perm]
  if !exists {
    caches = make(map[string]*ResourceAclCache)
    req.resCache[perm] = caches
  }
  cache, exists := caches[encodedResKey]
  if !exists {
    cache = NewResourceAclCache(resKey, perm)
    caches[encodedResKey] = cache
  }
  return cache
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
func (res *ResourceAclCache) IsAuthorizedRequestor(req *datastore.Key) bool {
  _, exists := res.AuthorizedRequestorKeys[req.Encode()]
  return exists
}

func (req *RequestorAclCache) can(perm Permission, res *ResourceAclCache) bool {
  // Check if the user is an authorized requestor.
  if _, exists := res.AuthorizedRequestorKeys[req.EncodedUserKey]; exists {
    return true
  }
  
  // Check if the user is in a group that is an authorized requestor.
  for encodedRequestor := range req.GroupKeys {
    if _, exists := res.AuthorizedRequestorKeys[encodedRequestor]; exists {
      return true
    }
  }
  return false
}
func (req *RequestorAclCache) Can(
  c appengine.Context, perm Permission, resKey *datastore.Key) error {
  // Allow application admin to do anything.
  gaeUser := user.Current(c)
  if gaeUser != nil && gaeUser.Admin {
    return nil
  }
  
  res := req.lookupResourceAcls(perm, resKey)
  
  // Ensure ACL caches are initialized.
  if err := req.init(c); err != nil { return err }
  if err := res.init(c); err != nil { return err }
  
  if req.can(perm, res) {
    return nil
  } else {
    return ErrNotAuthorized{perm, resKey}
  }
}

func (req *RequestorAclCache) FindAll(
  c appengine.Context, resourceKind string, perm Permission) ([]*datastore.Key, error) {
  if err := req.init(c); err != nil { return nil, err }
  
  allKeySet := make(map[string]*datastore.Key)
  
  for _, groupKey := range req.GroupKeys {
    keys, err := AclFindAll(c, groupKey, resourceKind, perm)
    if err != nil { return nil, err }
    
    for _, key := range keys {
      allKeySet[key.Encode()] = key
    }
  }
  
  allKeys := make([]*datastore.Key, 0, len(allKeySet))
  for _, key := range allKeySet {
    allKeys = append(allKeys, key)
  }
  return allKeys, nil
}

// For all the groups this user belongs to, returns a permission map for the given resource.
func (req *RequestorAclCache) PermissionMapFor(
  c appengine.Context,
  resKey *datastore.Key) ([]*Group, []*datastore.Key, []map[Permission]bool, error) {
  if err := req.init(c); err != nil { return nil, nil, nil, err }
  
  resByPerm := make(map[Permission]*ResourceAclCache)
  for _, perm := range AllPermissions() {
    resByPerm[perm] = req.lookupResourceAcls(perm, resKey)
    if err := resByPerm[perm].init(c); err != nil { return nil, nil, nil, err }
  }
  
  groupKeys := make([]*datastore.Key, len(req.GroupKeys))
  groups := make([]*Group, len(groupKeys))
  maps := make([]map[Permission]bool, len(groupKeys))
  
  i := 0
  for v, key := range req.GroupKeys {
    groupKeys[i] = key
    groups[i] = req.Groups[v]
    maps[i] = make(map[Permission]bool)
    for _, perm := range AllPermissions() {
      maps[i][perm] = resByPerm[perm].IsAuthorizedRequestor(groupKeys[i])
    }
    i++
  }
  return groups, groupKeys, maps, nil
}