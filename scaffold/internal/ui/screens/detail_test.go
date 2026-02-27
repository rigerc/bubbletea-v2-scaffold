package screens

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"scaffold/internal/task"
)

// newLoadingDetail returns a Detail with loading already active (simulating
// Init() having been called without running the program loop).
func newLoadingDetail(t *testing.T) *Detail {
	t.Helper()
	d := NewDetail("title", "description", "test-id", context.Background())
	d.load.Start()
	return d
}

// --- DoneMsg ---

func TestDetail_DoneMsg_StopsLoading(t *testing.T) {
	d := newLoadingDetail(t)
	assert.True(t, d.load.Active(), "precondition: loading must be active")

	m, cmd := d.Update(task.DoneMsg[string]{Label: "detail-load", Value: "loaded"})

	detail := m.(*Detail)
	assert.False(t, detail.load.Active(), "loading should stop after DoneMsg")
	assert.Nil(t, cmd, "no follow-up command expected")
}

func TestDetail_DoneMsg_WrongLabel_IgnoresMessage(t *testing.T) {
	d := newLoadingDetail(t)

	m, _ := d.Update(task.DoneMsg[string]{Label: "other-task", Value: "x"})

	detail := m.(*Detail)
	assert.True(t, detail.load.Active(), "loading should still be active for unrelated label")
}

// --- ErrMsg ---

func TestDetail_ErrMsg_StopsLoading(t *testing.T) {
	d := newLoadingDetail(t)

	m, cmd := d.Update(task.ErrMsg{Label: "detail-load", Err: errors.New("timeout")})

	detail := m.(*Detail)
	assert.False(t, detail.load.Active(), "loading should stop after ErrMsg")
	assert.Nil(t, cmd, "no follow-up command expected")
}

func TestDetail_ErrMsg_WrongLabel_IgnoresMessage(t *testing.T) {
	d := newLoadingDetail(t)

	m, _ := d.Update(task.ErrMsg{Label: "other-task", Err: errors.New("x")})

	detail := m.(*Detail)
	assert.True(t, detail.load.Active(), "loading should still be active for unrelated label")
}

// --- Esc key ---

func TestDetail_EscKey_SendsBackMsg(t *testing.T) {
	d := NewDetail("title", "desc", "id", context.Background())
	// Loading is not active: Esc should produce a BackMsg.

	_, cmd := d.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.NotNil(t, cmd, "Esc should return a command")

	msg := cmd()
	_, ok := msg.(BackMsg)
	assert.True(t, ok, "command should produce a BackMsg")
}

func TestDetail_EscKey_WhileLoading_IsConsumedBySpinner(t *testing.T) {
	d := newLoadingDetail(t)

	// While loading, key presses are forwarded to the spinner, not to the
	// Esc handler, so no BackMsg is sent.
	_, cmd := d.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		msg := cmd()
		_, isBack := msg.(BackMsg)
		assert.False(t, isBack, "BackMsg should not be sent while loading")
	}
}

// --- Tick ---

func TestDetail_Tick_IncrementsElapsed(t *testing.T) {
	d := newLoadingDetail(t)
	assert.Equal(t, 0, d.elapsed, "precondition: no time elapsed yet")

	now := time.Now()
	m, cmd := d.Update(detailTickMsg(now))

	detail := m.(*Detail)
	assert.Equal(t, 1, detail.elapsed, "elapsed should increment on tick")
	assert.NotNil(t, cmd, "tick should reschedule itself while loading")
}

func TestDetail_Tick_StopsAfterLoadDone(t *testing.T) {
	d := newLoadingDetail(t)

	// Stop loading, then send a tick.
	d.load.Stop()
	m, cmd := d.Update(detailTickMsg(time.Now()))

	detail := m.(*Detail)
	assert.Equal(t, 0, detail.elapsed, "elapsed should not increment after loading stops")
	assert.Nil(t, cmd, "no reschedule after loading stops")
}

// --- Body ---

func TestDetail_Body_ShowsElapsedWhileLoading(t *testing.T) {
	d := newLoadingDetail(t)
	d.elapsed = 2

	body := d.Body()
	assert.Contains(t, body, "2s", "body should display elapsed seconds while loading")
}

func TestDetail_Body_ShowsContentAfterLoad(t *testing.T) {
	d := NewDetail("My Title", "My Desc", "screen-id", context.Background())
	// load is not active (Init not called)

	body := d.Body()
	assert.Contains(t, body, "My Title")
	assert.Contains(t, body, "screen-id")
}
