package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

// Permutations holds all tasks for permutation in progress
type Permutations struct {
	lastID uint64
	jobs   map[string]chan []int
	mu     *sync.RWMutex
}

func (p *Permutations) nextID() string {
	return strconv.FormatUint(atomic.AddUint64(&p.lastID, 1), 10)
}

func (p *Permutations) setChan(id string, ch chan []int) {
	p.mu.Lock()
	p.jobs[id] = ch
	p.mu.Unlock()
}

func (p *Permutations) getChan(id string) (ch chan []int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.jobs[id]
}

// POST /api/v1/init
func (p *Permutations) initCtrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(http.StatusBadRequest, "Bad data", w)
		return
	}

	set := []int{}
	if err = json.Unmarshal(data, &set); err != nil {
		ErrorJSON(http.StatusBadRequest, "Bad data", w)
		return
	}
	if !isValidSet(set) {
		ErrorJSON(http.StatusBadRequest, "Invalid set", w)
		return
	}

	// prepare set for permutation algorithm
	sort.SliceStable(set, func(i, j int) bool { return set[i] < set[j] })

	jobID := p.nextID()
	ch := make(chan []int)
	p.setChan(jobID, ch)

	go func() {
		n := len(set)
		for permutate(set, n) {
			ch <- set
		}
		close(ch)
	}()

	resp := map[string]interface{}{"jobID": jobID, "success": true}
	respB, err := json.Marshal(resp)
	if err != nil {
		ErrorJSON(http.StatusInternalServerError, "Server problem", w)
		return
	}
	SuccessJSON(respB, w)
}

// GET /api/v1/next
// Clients must send x-job-id header
func (p *Permutations) nextCtrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	jobID := r.Header.Get("X-JOB-ID")
	ch := p.getChan(jobID)
	if ch == nil {
		ErrorJSON(http.StatusNotFound, "Job does not exist", w)
		return
	}

	data := <-ch
	if data == nil {
		data = []int{}
	}
	b, err := json.Marshal(data)
	if err != nil {
		ErrorJSON(http.StatusBadRequest, "Bad data", w)
		return
	}
	SuccessJSON(b, w)
}

func processData(data []int, ch chan []int) {
	n := len(data)
	for permutate(data, n) {
		ch <- data
	}
	close(ch)
}

func isValidSet(data []int) bool {
	set := map[int]struct{}{}
	for _, v := range data {
		if v < 0 {
			return false
		}
		if _, ok := set[v]; ok {
			return false
		}
		set[v] = struct{}{}
	}
	return true
}

func permutate(data []int, n int) bool {
	i := n - 2
	for i >= 0 && data[i] > data[i+1] {
		i--
	}
	if i == -1 {
		return false
	}
	for j, k := i+1, n-1; j < k; j, k = j+1, k-1 {
		data[j], data[k] = data[k], data[j]
	}
	j := i + 1
	for data[j] < data[i] {
		j++
	}
	data[i], data[j] = data[j], data[i]
	return true
}
