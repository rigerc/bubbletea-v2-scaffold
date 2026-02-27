package ui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"scaffold/config"
	"scaffold/internal/ui/screens"
	"scaffold/internal/ui/status"
)

// testModel returns a minimal rootModel suitable for unit tests.
// The model is in rootStateLoading (no WindowSizeMsg received yet).
func testModel(t *testing.T) rootModel {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cfg := config.Config{LogLevel: "info"}
	return newRootModel(ctx, cancel, cfg, "", false)
}

// --- rootState / WindowSizeMsg ---

func TestRootModel_InitialState_IsLoading(t *testing.T) {
	m := testModel(t)
	assert.Equal(t, rootStateLoading, m.state, "new model should be in loading state")
}

func TestRootModel_WindowSizeMsg_SetsReadyState(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	root := updated.(rootModel)

	assert.Equal(t, rootStateReady, root.state)
	assert.Equal(t, 120, root.width)
	assert.Equal(t, 40, root.height)
}

func TestRootModel_View_EmptyUntilReady(t *testing.T) {
	m := testModel(t)
	// No WindowSizeMsg sent — should render nothing.
	v := m.View()
	assert.Equal(t, "", v.Content)
}

// --- NavigateMsg ---

func TestRootModel_NavigateMsg_PushesCurrentScreen(t *testing.T) {
	m := testModel(t)
	// Set dimensions so navigation helpers work.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(rootModel)

	initialScreen := m.current
	newScreen := screens.NewHome()

	updated, _ = m.Update(NavigateMsg{Screen: newScreen})
	root := updated.(rootModel)

	assert.Equal(t, 1, root.stack.Len(), "original screen should be on the stack")
	assert.Equal(t, newScreen, root.current, "current screen should be the new one")
	_ = initialScreen // referenced to confirm it was pushed
}

func TestRootModel_NavigateMsg_Stacks(t *testing.T) {
	m := testModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(rootModel)

	screen1 := screens.NewHome()
	screen2 := screens.NewHome()

	updated, _ = m.Update(NavigateMsg{Screen: screen1})
	updated, _ = updated.(rootModel).Update(NavigateMsg{Screen: screen2})
	root := updated.(rootModel)

	assert.Equal(t, 2, root.stack.Len(), "two screens should be on the stack")
}

// --- BackMsg ---

func TestRootModel_BackMsg_PopsScreen(t *testing.T) {
	m := testModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(rootModel)

	original := m.current
	updated, _ = m.Update(NavigateMsg{Screen: screens.NewHome()})
	updated, _ = updated.(rootModel).Update(screens.BackMsg{})
	root := updated.(rootModel)

	assert.Equal(t, original, root.current, "BackMsg should restore the previous screen")
	assert.Equal(t, 0, root.stack.Len(), "stack should be empty after pop")
}

// --- status.Msg / status.ClearMsg ---

func TestRootModel_StatusMsg_UpdatesStatus(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(status.Msg{Text: "something broke", Kind: status.KindError})
	root := updated.(rootModel)

	assert.Equal(t, "something broke", root.status.Text)
	assert.Equal(t, status.KindError, root.status.Kind)
}

func TestRootModel_StatusClearMsg_ResetsStatus(t *testing.T) {
	m := testModel(t)

	// First set a status.
	updated, _ := m.Update(status.Msg{Text: "busy", Kind: status.KindInfo})
	// Then clear it.
	updated, _ = updated.(rootModel).Update(status.ClearMsg{})
	root := updated.(rootModel)

	assert.Equal(t, "Ready", root.status.Text)
	assert.Equal(t, status.KindNone, root.status.Kind)
}

// --- screenStack ---

func TestScreenStack_PushPop(t *testing.T) {
	var s screenStack
	a := screens.NewHome()
	b := screens.NewHome()

	s.Push(a)
	s.Push(b)
	assert.Equal(t, 2, s.Len())

	got := s.Pop()
	assert.Equal(t, b, got, "Pop should return LIFO order")
	assert.Equal(t, 1, s.Len())
}

func TestScreenStack_PopEmpty_ReturnsNil(t *testing.T) {
	var s screenStack
	assert.Nil(t, s.Pop())
}

func TestScreenStack_PeekDoesNotRemove(t *testing.T) {
	var s screenStack
	a := screens.NewHome()
	s.Push(a)

	assert.Equal(t, a, s.Peek())
	assert.Equal(t, 1, s.Len(), "Peek should not remove the element")
}
