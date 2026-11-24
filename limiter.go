package limiter
import ("sync"; "time")
type bucket struct { tokens float64; updated time.Time }
type Limiter struct { mu sync.Mutex; buckets map[string]bucket; rate, capacity float64; now func() time.Time }
func New(rate float64, capacity int) *Limiter { if rate<=0 || capacity<=0 { panic("rate and capacity must be positive") }; return &Limiter{buckets:make(map[string]bucket),rate:rate,capacity:float64(capacity),now:time.Now} }
func (l *Limiter) Allow(key string) bool { l.mu.Lock(); defer l.mu.Unlock(); now:=l.now(); b,ok:=l.buckets[key]; if !ok { b=bucket{tokens:l.capacity,updated:now} }; elapsed:=now.Sub(b.updated).Seconds(); if elapsed>0 { b.tokens+=elapsed*l.rate; if b.tokens>l.capacity { b.tokens=l.capacity }; b.updated=now }; allowed:=b.tokens>=1; if allowed { b.tokens-- }; l.buckets[key]=b; return allowed }
