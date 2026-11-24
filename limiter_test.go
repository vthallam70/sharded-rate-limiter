package limiter
import ("testing"; "time")
func TestRefillAndIsolation(t *testing.T) { l:=New(1,1); now:=time.Unix(0,0); l.now=func() time.Time{return now}; if !l.Allow("a") || l.Allow("a") || !l.Allow("b") {t.Fatal("capacity or isolation failed")}; now=now.Add(time.Second); if !l.Allow("a") {t.Fatal("refill failed")} }
func BenchmarkAllow(b *testing.B) { l:=New(1000000,1000000); b.RunParallel(func(pb *testing.PB){for pb.Next(){l.Allow("user")}}) }
