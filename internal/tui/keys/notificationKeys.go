package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type NotificationKeyMap struct {
	View                 key.Binding
	BackToNotification   key.Binding
	MarkAsDone           key.Binding
	MarkAllAsDone        key.Binding
	MarkAsRead           key.Binding
	MarkAllAsRead        key.Binding
	Unsubscribe          key.Binding
	ToggleBookmark       key.Binding
	Open                 key.Binding
	SortByRepo           key.Binding
	SwitchToPRs          key.Binding
	ToggleSmartFiltering key.Binding
}

var NotificationKeys = NotificationKeyMap{
	View: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view notification"),
	),
	BackToNotification: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back to notification"),
	),
	MarkAsDone: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "mark as done"),
	),
	MarkAllAsDone: key.NewBinding(
		key.WithKeys("alt+d"),
		key.WithHelp("Alt+d", "mark all as done"),
	),
	MarkAsRead: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "mark as read"),
	),
	MarkAllAsRead: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "mark all as read"),
	),
	Unsubscribe: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "unsubscribe"),
	),
	ToggleBookmark: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "toggle bookmark"),
	),
	Open: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in browser"),
	),
	SortByRepo: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "sort by repo"),
	),
	SwitchToPRs: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch to PRs"),
	),
	ToggleSmartFiltering: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle smart filtering"),
	),
}

func NotificationFullHelp() []key.Binding {
	return []key.Binding{
		NotificationKeys.View,
		NotificationKeys.BackToNotification,
		NotificationKeys.MarkAsDone,
		NotificationKeys.MarkAllAsDone,
		NotificationKeys.MarkAsRead,
		NotificationKeys.MarkAllAsRead,
		NotificationKeys.Unsubscribe,
		NotificationKeys.ToggleBookmark,
		NotificationKeys.Open,
		NotificationKeys.SortByRepo,
		NotificationKeys.SwitchToPRs,
		NotificationKeys.ToggleSmartFiltering,
	}
}
