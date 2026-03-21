package constants

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	FirstItem     key.Binding
	LastItem      key.Binding
	TogglePreview key.Binding
	OpenGithub    key.Binding
	Refresh       key.Binding
	PageDown      key.Binding
	PageUp        key.Binding
	NextSection   key.Binding
	PrevSection   key.Binding
	Help          key.Binding
	Quit          key.Binding
}

type Dimensions struct {
	Width  int
	Height int
}

const (
	Ellipsis = "…"

	ApprovedIcon         = "󰄬"
	ChangesRequestedIcon = ""
	DotIcon              = ""
	SmallDotIcon         = "⋅"
	HorizontalLineIcon   = "─"
	EmptyIcon            = ""
	FailureIcon          = "󰅙"
	PersonIcon           = ""
	SuccessIcon          = ""
	TeamIcon             = ""
	WaitingIcon          = ""
	ActionRequiredIcon   = "" // nf-cod-warning (matches GitHub UI)

	BehindIcon         = "󰇮"
	BlockedIcon        = ""
	ClosedIcon         = ""
	CodeReviewIcon     = ""
	CommentIcon        = ""
	CommentsIcon       = ""
	DonateIcon         = "󱃱"
	DraftIcon          = ""
	CommitIcon         = ""
	VerticalCommitIcon = "󰜘"
	LabelsIcon         = "󰌖"
	MergedIcon         = ""
	MergeQueueIcon     = "" // \uf4db nf-oct-git_merge_queue
	OpenIcon           = ""
	SelectionIcon      = "❯"

	AutocompleteColumnGap              = 2
	AutocompleteMinValueWidth          = 8
	AutocompleteMinDetailWidth         = 10
	AutocompletePreferredValueRatioNum = 2
	AutocompletePreferredValueRatioDen = 3

	// New contributors: users who created a PR for the repo for the first time
	NewContributorIcon = "󰎔" // \udb80\udf94 nf-md-new_box

	// Contributors: everyone who has contributed something back to the project
	ContributorIcon = "" // \uedc6 nf-fa-user_check

	// Collaborator is a person who isn't explicitly a member of your organization,
	// but who has Read, Write, or Admin permissions to one or more repositories in your organization.
	CollaboratorIcon = "" // \uedcf nf-fa-user_shield

	// A member of the organization
	MemberIcon = "" // \uf42b nf-oct-organization

	// The person/s who has administrative ownership over the organization or repository (not always the same as the original author)
	OwnerIcon = "" // \uf511 nf-oct-shield_lock

	UnknownRoleIcon = "󰭙" // \udb82\udf59 nf-md-account_question

	// Notification type icons
	WorkflowIcon     = "" // \uf52e nf-oct-checklist (for CheckSuite/CI)
	WorkflowRunIcon  = "" // \uebd6 nf-cod-workflow (for CheckSuite default)
	SecurityIcon     = "󰒃" // \udb80\udc83 nf-md-shield_alert (for security alerts)
	NotificationIcon = "" // \ueaa2 nf-cod-bell (generic notification fallback)
	SearchIcon       = "" // \uf002 nf-fa-search

	// Prompts
	AssignPrompt   = "Assign users (whitespace-separated)" + Ellipsis
	UnassignPrompt = "Unassign users (whitespace-separated)" + Ellipsis
	CommentPrompt  = "Leave a comment" + Ellipsis
	ApprovalPrompt = "Approve with comment" + Ellipsis
	LabelPrompt    = "Add/remove labels (comma-separated)" + Ellipsis

	Logo = `▜▔▚▐▔▌▚▔▐ ▌
▟▁▞▐▔▌▁▚▐▔▌`
)
