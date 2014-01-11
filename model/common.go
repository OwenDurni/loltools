package model

import (
  "appengine"
  "appengine/datastore"
  "encoding/base64"
  "encoding/binary"
  "errors"
)

func EncodeGlobalKeyShort(k *datastore.Key) string {
  buf := make([]byte, 8)
  binary.PutVarint(buf, k.IntID())
  return base64.URLEncoding.EncodeToString(buf)
}

func DecodeGlobalKeyShort(
    c appengine.Context, kind string, encodedKey string) (*datastore.Key, error) {
  buf, err := base64.URLEncoding.DecodeString(encodedKey)
  if err != nil {
    return nil, err
  }
  id, n := binary.Varint(buf)
  if n == 0 {
    return nil, errors.New("DecodeGlobalKeyShort(): buf too small")
  } else if n < 0 {
    return nil, errors.New("DecodeGlobalKeyShort(): overflow")
  }
  return datastore.NewKey(c, kind, "", id, nil), nil
}