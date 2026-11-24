package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	limiter "github.com/vthallam70/sharded-rate-limiter"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

type server struct {
	self    string
	nodes   []string
	limiter *limiter.Limiter
	client  *http.Client
}

func owner(key string, nodes []string) string {
	var best string
	var max uint64
	for _, n := range nodes {
		h := sha256.Sum256([]byte(key + "\x00" + n))
		score := binary.BigEndian.Uint64(h[:8])
		if best == "" || score > max {
			best = n
			max = score
		}
	}
	return best
}
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/allow" || r.Method != "POST" {
		http.Error(w, "POST /allow", 404)
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req) != nil || req.Key == "" || len(req.Key) > 256 {
		http.Error(w, "invalid key", 400)
		return
	}
	target := owner(req.Key, s.nodes)
	if target != s.self {
		if r.Header.Get("X-Forwarded-Limiter") == "1" {
			http.Error(w, "inconsistent cluster configuration", 503)
			return
		}
		body := strings.NewReader(fmt.Sprintf(`{"key":%q}`, req.Key))
		upstream, err := http.NewRequestWithContext(r.Context(), "POST", target+"/allow", body)
		if err != nil {
			http.Error(w, "invalid upstream", 500)
			return
		}
		upstream.Header.Set("X-Forwarded-Limiter", "1")
		resp, err := s.client.Do(upstream)
		if err != nil {
			http.Error(w, "owner unavailable; request denied", 503)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, io.LimitReader(resp.Body, 4096))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"allowed": s.limiter.Allow(req.Key)})
}
func main() {
	addr := flag.String("listen", ":8080", "listen address")
	self := flag.String("self", "http://127.0.0.1:8080", "public node URL")
	peers := flag.String("nodes", "http://127.0.0.1:8080", "comma-separated node URLs")
	rate := flag.Float64("rate", 1, "tokens per second")
	capacity := flag.Int("capacity", 10, "bucket capacity")
	flag.Parse()
	nodes := strings.Split(*peers, ",")
	sort.Strings(nodes)
	found := false
	for _, n := range nodes {
		if n == *self {
			found = true
		}
	}
	if !found {
		log.Fatal("self must appear in nodes")
	}
	s := &server{self: *self, nodes: nodes, limiter: limiter.New(*rate, *capacity), client: &http.Client{Timeout: 2 * time.Second}}
	log.Fatal(http.ListenAndServe(*addr, s))
}
