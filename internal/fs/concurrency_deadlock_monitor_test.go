package fs

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type workerState struct {
	id         int
	lastAction atomic.Value
	lastDetail atomic.Value
	lastUpdate int64
	lastAlert  int64
}

type deadlockMonitor struct {
	t            *testing.T
	testName     string
	threshold    time.Duration
	states       []*workerState
	traceEnabled bool
	stackEnabled bool
	stopCh       chan struct{}
	watchWg      sync.WaitGroup
}

func newDeadlockMonitor(t *testing.T, testName string, workerCount int) *deadlockMonitor {
	traceRaw := strings.ToLower(os.Getenv("DEADLOCK_TRACE"))
	monitor := &deadlockMonitor{
		t:            t,
		testName:     testName,
		threshold:    time.Second,
		states:       make([]*workerState, workerCount),
		traceEnabled: traceRaw != "",
		stackEnabled: strings.Contains(traceRaw, "stack"),
		stopCh:       make(chan struct{}),
	}
	now := time.Now().UnixNano()
	for i := range monitor.states {
		state := &workerState{id: i}
		state.lastAction.Store("init")
		state.lastDetail.Store("")
		atomic.StoreInt64(&state.lastUpdate, now)
		monitor.states[i] = state
	}
	if monitor.traceEnabled {
		monitor.watchWg.Add(1)
		go monitor.watch()
	}
	return monitor
}

func (m *deadlockMonitor) watch() {
	defer m.watchWg.Done()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.scan()
		}
	}
}

func (m *deadlockMonitor) scan() {
	if len(m.states) == 0 {
		return
	}
	now := time.Now()
	stalled := false
	for _, state := range m.states {
		last := atomic.LoadInt64(&state.lastUpdate)
		if last == 0 {
			continue
		}
		if now.Sub(time.Unix(0, last)) > m.threshold {
			prev := atomic.LoadInt64(&state.lastAlert)
			if prev != 0 && now.Sub(time.Unix(0, prev)) < m.threshold {
				continue
			}
			atomic.StoreInt64(&state.lastAlert, now.UnixNano())
			stalled = true
		}
	}
	if stalled {
		m.Snapshot(fmt.Sprintf("workers stalled > %s", m.threshold), m.stackEnabled)
	}
}

func (m *deadlockMonitor) Record(workerID int, action string, detail ...string) {
	if workerID < 0 || workerID >= len(m.states) {
		return
	}
	state := m.states[workerID]
	state.lastAction.Store(action)
	d := ""
	if len(detail) > 0 {
		d = detail[0]
	}
	state.lastDetail.Store(d)
	atomic.StoreInt64(&state.lastUpdate, time.Now().UnixNano())
}

func (m *deadlockMonitor) Snapshot(reason string, includeStacks bool) {
	var builder strings.Builder
	now := time.Now()
	builder.WriteString(fmt.Sprintf("[deadlock-monitor] %s | test=%s | goroutines=%d\n",
		reason, m.testName, runtime.NumGoroutine()))
	for _, state := range m.states {
		last := time.Unix(0, atomic.LoadInt64(&state.lastUpdate))
		action, _ := state.lastAction.Load().(string)
		detail, _ := state.lastDetail.Load().(string)
		if detail != "" {
			action = action + " " + detail
		}
		builder.WriteString(fmt.Sprintf("worker=%02d action=%q ago=%s\n",
			state.id, action, now.Sub(last).Truncate(time.Millisecond)))
	}
	m.t.Log(builder.String())
	if includeStacks {
		var buf bytes.Buffer
		if prof := pprof.Lookup("goroutine"); prof != nil {
			_ = prof.WriteTo(&buf, 2)
		}
		if buf.Len() > 0 {
			m.t.Logf("[deadlock-monitor] goroutine dump:\n%s", buf.String())
		}
	}
}

func (m *deadlockMonitor) Wait(wg *sync.WaitGroup, timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		m.Snapshot(fmt.Sprintf("workers failed to finish within %s", timeout), true)
		return fmt.Errorf("workers did not finish within %s", timeout)
	}
}

func (m *deadlockMonitor) Stop() {
	if m.traceEnabled {
		close(m.stopCh)
		m.watchWg.Wait()
	}
}

func (m *deadlockMonitor) ForceSnapshot(reason string) {
	m.Snapshot(reason, m.stackEnabled)
}
