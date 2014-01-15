package model

import (
  "appengine"
  "appengine/memcache"
  "errors"
  "fmt"
  "math"
  "time"
)

// Describes a rate limit that permits MaxEvents in the last IntervalSeconds.
// Ex: RateLimit{10, 100} would indicate 10 events per 100 seconds.
type RateLimit struct {
  MaxEvents       int
  IntervalSeconds int
}

type DistributedRateLimiter struct {
  Name   string
  Limits []RateLimit
}

// Creates a new DistributedRateLimiter that allows events according to the
// specified limits (or re-initializes the limits of an existing rate limiter
// with the same name).
func (r DistributedRateLimiter) Init(c appengine.Context) error {
  e := new(DistributedRateLimiterEntity)
  e.Buckets = make([]TokenBucket, len(r.Limits))
  for i, _ := range e.Buckets {
    e.Buckets[i].SetLimit(r.Limits[i])
  }
  
  err := memcache.JSON.Set(c, &memcache.Item{
    Key: fmt.Sprintf("DistributedRateLimiterEntity/%s", r.Name),
    Object: e,
  })
  return err
}

// Consumes tokens from the rate limiter if nil is returned.
// Otherwise an error is returned describing why tokens could not
// be consumed.
func (r *DistributedRateLimiter) TryConsume(c appengine.Context, events int) error {
  e := new(DistributedRateLimiterEntity)
  key := fmt.Sprintf("DistributedRateLimiterEntity/%s", r.Name)
  for attempt := 0; attempt < 10; attempt++ {
    item, err := memcache.JSON.Get(c, key, e)
    if err != nil {
      if err == memcache.ErrCacheMiss {
        // Put a fresh item in memcache and try again.
        r.Init(c)
        continue
      } else {
        return err
      }
    }
    
    e.addTokens(time.Now().UTC())
    
    if err := e.tryConsume(events); err != nil {
      // We are being rate limited.
      return err
    }
    
    // We've consumed the tokens if and only if nothing else has written the
    // entry back to memcache since we got it.
    item.Object = e
    if err := memcache.JSON.CompareAndSwap(c, item); err != nil {
      continue
    }
    return nil
  }
  return errors.New(fmt.Sprintf("RetryLimit for memcache.CompareAndSwap reached: %s", key))
}

type TokenBucket struct {
  Limit RateLimit

  // The number of tokens in the bucket at LastCheckTime.
  Tokens float64

  // Tracks the last time this token bucket was processed so that we can
  // optimize adding tokens to the bucket each update instead of every
  // Rate seconds. Time is in UTC.
  LastCheckTime time.Time
}

// Computes the number of tokens to add to the bucket per second.
func (b *TokenBucket) newTokensPerSecond() float64 {
  return float64(b.Limit.MaxEvents) / float64(b.Limit.IntervalSeconds)
}

func (b *TokenBucket) AddTokens(t time.Time) {
  elapsedSeconds := t.Sub(b.LastCheckTime).Seconds()
  if elapsedSeconds < 0 {
    return
  }
  b.Tokens = math.Min(b.Tokens+elapsedSeconds*b.newTokensPerSecond(),
    float64(b.Limit.MaxEvents))
  b.LastCheckTime = t
}

// If the provided limits are the same as the existing limits this is a no-op.
// Otherwise the limits are set to the new limits and all tokens are removed from
// the bucket.
func (b *TokenBucket) SetLimit(limit RateLimit) {
  if limit == b.Limit {
    return
  }
  b.Limit = limit
  b.Tokens = 0.0
  b.LastCheckTime = time.Now().UTC()
}

// An opaque type for managing distributed rate limits with the appengine datastore.
// Use its methods to interact with it.
type DistributedRateLimiterEntity struct {
  // Counts of the total number of requests accepted.
  AcceptCount int64

  // Internal buckets.
  Buckets []TokenBucket
}

func (e *DistributedRateLimiterEntity) addTokens(t time.Time) {
  for i, _ := range e.Buckets {
    e.Buckets[i].AddTokens(t)
  }
}

func (e *DistributedRateLimiterEntity) tryConsume(numTokens int) error {
  tokens := float64(numTokens)

  // See if tokens are available from all buckets before consuming any tokens.
  for _, b := range e.Buckets {
    if tokens > b.Tokens {
      return errors.New(fmt.Sprintf("Exceeded rate limit: %+v", b.Limit))
    }
  }
  // Consume the tokens.
  for i, _ := range e.Buckets {
    e.Buckets[i].Tokens -= tokens
  }
  e.AcceptCount++
  return nil
}
