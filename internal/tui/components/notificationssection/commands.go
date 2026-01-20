package notificationssection

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
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

// markAllAsDone marks all currently visible notifications in this section as done.
// "All" refers to the notifications currently loaded in m.Notifications, not all
// notifications on GitHub.
func (m *Model) markAllAsDone() tea.Cmd {
	if len(m.Notifications) == 0 {
		return nil
	}

	count := len(m.Notifications)
	taskId := "notification_done_all"
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Marking %d notifications as done", count),
		FinishedText: fmt.Sprintf("%d notifications marked as done", count),
		State:        context.TaskStart,
		Error:        nil,
	}

	notificationIds := make([]string, 0, count)
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

// openInBrowser marks the current notification as read and opens it in the browser
func (m *Model) openInBrowser() tea.Cmd {
	notification := m.GetCurrNotification()
	if notification == nil {
		return nil
	}

	notificationId := notification.GetId()
	notificationUrl := notification.GetUrl()

	return tea.Batch(
		func() tea.Msg {
			_ = data.MarkNotificationRead(notificationId)
			return UpdateNotificationReadStateMsg{
				Id:     notificationId,
				Unread: false,
			}
		},
		func() tea.Msg {
			b := browser.New("", os.Stdout, os.Stdin)
			err := b.Browse(notificationUrl)
			if err != nil {
				return constants.ErrMsg{Err: err}
			}
			return nil
		},
	)
}

// CheckoutPR checks out a PR. This is a standalone function that can be called
// from ui.go with the PR details from the notification view.
func CheckoutPR(ctx *context.ProgramContext, prNumber int, repoName string) (tea.Cmd, error) {
	repoPath, ok := common.GetRepoLocalPath(repoName, ctx.Config.RepoPaths)
	if !ok {
		return nil, errors.New("local path to repo not specified, set one in your config.yml under repoPaths")
	}

	taskId := fmt.Sprintf("checkout_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Checking out PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been checked out at %s", prNumber, repoPath),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command(
			"gh",
			"pr",
			"checkout",
			fmt.Sprint(prNumber),
		)
		userHomeDir, _ := os.UserHomeDir()
		if strings.HasPrefix(repoPath, "~") {
			repoPath = strings.Replace(repoPath, "~", userHomeDir, 1)
		}

		c.Dir = repoPath
		err := c.Run()
		return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
	}), nil
}
