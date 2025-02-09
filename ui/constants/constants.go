package constants

import (
	"github.com/charmbracelet/bubbles/key"
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
	WaitingIcon = ""
	FailureIcon = "󰅙"
	SuccessIcon = ""

	CommentIcon = ""
	DraftIcon   = ""
	BehindIcon  = "󰇮"
	BlockedIcon = ""
	MergedIcon  = ""
	OpenIcon    = ""
	ClosedIcon  = ""

	NewContributorIcon = "" // \uebe9 nf-cod-verified_filled
	ContributorIcon    = "" // \uedc6 nf-fa-user_check
	CollaboratorIcon   = "" // \uedcf nf-fa-user_shield
	MemberIcon         = "󰢏" // \udb82\udc8f nf-md-shield_account
	OwnerIcon          = "󱇐" // \udb84\uddd0 nf-md-crown_outline
	UnknownRoleIcon    = "󱐡" // \udb85\udc21 nf-md-incognito_circle
)
