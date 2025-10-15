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
	Ellipsis = "…"

	PersonIcon  = ""
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
	DonateIcon  = "󱃱"

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
	OwnerIcon = "󱇐" // \udb84\uddd0 nf-md-crown_outline

	UnknownRoleIcon = "󱐡" // \udb85\udc21 nf-md-incognito_circle

	Logo = `▜▔▚▐▔▌▚▔▐ ▌
▟▁▞▐▔▌▁▚▐▔▌`
)
