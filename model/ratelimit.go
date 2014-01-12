package model

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
  "math"
  "time"
)

// Describes a rate limit that permits MaxEvents in the last IntervalSeconds.
// Ex: RateLimit{10, 100} would indicate 10 events per 100 seconds.
type RateLimit struct {
  MaxEvents int
  IntervalSeconds int
}

type tokenBucket struct {
  Limit RateLimit
  
  // The number of tokens in the bucket at LastCheckTime.
  Tokens float64
  
  // Tracks the last time this token bucket was processed so that we can
  // optimize adding tokens to the bucket each update instead of every
  // Rate seconds. Time is in UTC.
  LastCheckTime time.Time
}

// Computes the number of tokens to add to the bucket per second.
func (b *tokenBucket) NewTokensPerSecond() float64 {
  return float64(b.Limit.MaxEvents) / float64(b.Limit.IntervalSeconds)
}

func (b *tokenBucket) AddTokens(t time.Time) {
  elapsedSeconds := t.Sub(b.LastCheckTime).Seconds()
  if elapsedSeconds < 0 { return }
  b.Tokens = math.Max(b.Tokens + elapsedSeconds * b.NewTokensPerSecond(),
                      float64(b.Limit.MaxEvents))
  b.LastCheckTime = t
}

// If the provided limits are the same as the existing limits this is a no-op.
// Otherwise the limits are set to the new limits and all tokens are removed from
// the bucket.
func (b *tokenBucket) SetLimit(limit RateLimit) {
  if limit == b.Limit {
    return
  }
  b.Limit = limit
  b.Tokens = 0.0
  b.LastCheckTime = time.Now().UTC()
}

type DistributedRateLimiter struct {
  Name string
}

// An opaque type for managing distributed rate limits with the appengine datastore.
// Use its methods to interact with it.
type DistributedRateLimiterEntity struct {  
  // Counts of the total number of requests accepted and rejected.
  AcceptCount int64
  RejectCount int64
  
  // Internal buckets.
  buckets []tokenBucket
}

func (e *DistributedRateLimiterEntity) AddTokens(t time.Time) {
  for i, _ := range e.buckets {
    e.buckets[i].AddTokens(t)
  }
}

func (e *DistributedRateLimiterEntity) TryConsume(numTokens int) error {
  tokens := float64(numTokens)
  
  // See if tokens are available from all buckets before consuming any tokens.
  for _, b := range e.buckets {
    if tokens > b.Tokens {
      return errors.New(fmt.Sprintf("Exceeded rate limit: %+v", b.Limit))
    }
  }
  // Consume the tokens.
  for i, _ := range e.buckets {
    e.buckets[i].Tokens -= tokens
  }
  return nil
}

// Creates a new DistributedRateLimiter that allows events according to the
// specified limits (or re-initializes the limits of an existing rate limiter
// with the same name).
func NewDistributedRateLimiter(
    c appengine.Context,
    name string,
    limits []RateLimit) (*DistributedRateLimiter, error) {
  r := new(DistributedRateLimiterEntity)
  key := datastore.NewKey(c, "DistributedRateLimiterEntity", name, 0, nil)
  
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    err := datastore.Get(c, key, r);
    if err != nil {
      if err != datastore.ErrNoSuchEntity {
        return err
      }
      r.buckets = make([]tokenBucket, len(limits))
      for i, limit := range limits {
        r.buckets[i].SetLimit(limit)
      }
    }
    
    // Truncate existing limits to desired length.
    if len(r.buckets) > len(limits) {
      r.buckets = r.buckets[:len(limits)]
    }
    
    // Ensure existing limits are correct.
    i := 0
    for ; i < len(r.buckets); i++ {
      r.buckets[i].SetLimit(limits[i])
    }
    
    // Append new limits.
    for ; i < len(limits); i++ {
      r.buckets = append(r.buckets, tokenBucket{})
      r.buckets[i].SetLimit(limits[i])
    }
    
    // Write any changes back to the datastore.
    _, err = datastore.Put(c, key, r)
    return err
  }, nil)
  if err != nil {
    return nil, err
  }
  ret := new(DistributedRateLimiter)
  ret.Name = name
  return ret, nil
}

// Consumes tokens from the rate limiter if nil is returned.
// Otherwise an error is returned describing why tokens could not
// be consumed.
func (r *DistributedRateLimiter) TryConsume(c appengine.Context, events int) error {
  key := datastore.NewKey(c, "DistributedRateLimiterEntity", r.Name, 0, nil)
  
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    var e DistributedRateLimiterEntity
    if err := datastore.Get(c, key, &e); err != nil {
      return err
    }
    
    // Add tokens since last attempt and try to consume ones for this attempt.
    e.AddTokens(time.Now().UTC())
    if err := e.TryConsume(events); err != nil {
      return errors.New(fmt.Sprintf("DistributedRateLimiter(%s): %s",
                                    r.Name, err.Error()))
    }
    
    // Write any changes back to the datastore.
    _, err := datastore.Put(c, key, e)
    return err
  }, nil)
  
  // The tokens were consumed if and only if the transaction was successful.
  return err
}