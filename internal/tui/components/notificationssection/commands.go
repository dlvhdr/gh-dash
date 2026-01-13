package notificationssection

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) markAsDone() tea.Cmd {
	notification := m.GetCurrNotification()
	if notification == nil {
		return nil
	}

	notificationId := notification.GetId()
	taskId := fmt.Sprintf("notification_done_%s", notificationId)
	task := context.Task{
		Id:           taskId,
		StartText:    "Marking notification as done",
		FinishedText: "Notification marked as done",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		err := data.MarkNotificationDone(notificationId)
		if err == nil {
			// Persist to done store so it stays hidden across sessions
			data.GetDoneStore().MarkDone(notificationId)
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: UpdateNotificationMsg{
				Id:        notificationId,
				IsRemoved: err == nil,
			},
		}
	})
}

func (m *Model) markAllAsDone() tea.Cmd {
	if len(m.Notifications) == 0 {
		return nil
	}

	taskId := "notification_done_all"
	task := context.Task{
		Id:           taskId,
		StartText:    "Marking all notifications as done",
		FinishedText: "All notifications marked as done",
		State:        context.TaskStart,
		Error:        nil,
	}

	notificationIds := make([]string, 0, len(m.Notifications))
	for _, n := range m.Notifications {
		notificationIds = append(notificationIds, n.GetId())
	}

	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		// Mark each notification as done (delete it)
		doneStore := data.GetDoneStore()
		var lastErr error
		for _, id := range notificationIds {
			if err := data.MarkNotificationDone(id); err != nil {
				lastErr = err
			} else {
				// Persist to done store so it stays hidden across sessions
				doneStore.MarkDone(id)
			}
		}

		if lastErr != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: SectionType,
				TaskId:      taskId,
				Err:         lastErr,
			}
		}

		// Clear all notifications after marking as done
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         nil,
			Msg:         ClearAllNotificationsMsg{},
		}
	})
}

func (m *Model) markAllAsRead() tea.Cmd {
	taskId := "notification_read_all"
	task := context.Task{
		Id:           taskId,
		StartText:    "Marking all notifications as read",
		FinishedText: "All notifications marked as read",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		err := data.MarkAllNotificationsRead()
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: SectionType,
				TaskId:      taskId,
				Err:         err,
			}
		}

		// Update all notifications to read state
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         nil,
			Msg:         MarkAllAsReadMsg{},
		}
	})
}

type (
	// RefetchNotificationsMsg signals that notifications should be refetched from the API
	RefetchNotificationsMsg struct{}
	// ClearAllNotificationsMsg signals that all notifications should be removed from the local list
	// This is sent after successfully marking all notifications as done
	ClearAllNotificationsMsg struct{}
	// MarkAllAsReadMsg signals that all notifications should be updated to read state in the UI
	// This is sent after successfully calling the mark-all-read API
	MarkAllAsReadMsg struct{}
)

func (m *Model) markAsRead() tea.Cmd {
	notification := m.GetCurrNotification()
	if notification == nil {
		return nil
	}

	notificationId := notification.GetId()
	taskId := fmt.Sprintf("notification_read_%s", notificationId)
	task := context.Task{
		Id:           taskId,
		StartText:    "Marking notification as read",
		FinishedText: "Notification marked as read",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		err := data.MarkNotificationRead(notificationId)
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: UpdateNotificationReadStateMsg{
				Id:     notificationId,
				Unread: false,
			},
		}
	})
}

func (m *Model) unsubscribe() tea.Cmd {
	notification := m.GetCurrNotification()
	if notification == nil {
		return nil
	}

	notificationId := notification.GetId()
	taskId := fmt.Sprintf("notification_unsubscribe_%s", notificationId)
	task := context.Task{
		Id:           taskId,
		StartText:    "Unsubscribing from thread",
		FinishedText: "Unsubscribed from thread",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		err := data.UnsubscribeFromThread(notificationId)
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: UnsubscribedMsg{
				Id: notificationId,
			},
		}
	})
}

// UnsubscribedMsg is sent when a notification thread is unsubscribed
type UnsubscribedMsg struct {
	Id string
}

// UpdateNotificationReadStateMsg is sent when a notification's read state changes
type UpdateNotificationReadStateMsg struct {
	Id     string
	Unread bool
}
