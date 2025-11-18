# Styling and Theming System

This document explains how styling and theming work in gh-dash, including colors, layout, and user customization.

## Table of Contents
1. [Styling Philosophy](#styling-philosophy)
2. [Lipgloss Fundamentals](#lipgloss-fundamentals)
3. [Theme System](#theme-system)
4. [Styles Structure](#styles-structure)
5. [Adaptive Colors](#adaptive-colors)
6. [Layout and Spacing](#layout-and-spacing)
7. [User Customization](#user-customization)
8. [Best Practices](#best-practices)

---

## Styling Philosophy

### Core Principles

1. **Declarative**: Styles defined like CSS, applied to content
2. **Composable**: Combine multiple styles
3. **Responsive**: Adapt to terminal size
4. **Themeable**: User-customizable colors
5. **Accessible**: Light and dark mode support

### Why Lipgloss?

```
Traditional Approach          →    Lipgloss Approach
--------------------               -------------------
printf("\033[31m%s\033[0m", text)  style.Render(text)

Hard to read                       Declarative
Error-prone                        Type-safe
Not reusable                       Reusable
Manual calculations                Automatic layout
```

---

## Lipgloss Fundamentals

### Basic Style Creation

```go
import "github.com/charmbracelet/lipgloss"

// Create a style
var style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("205")).
    Background(lipgloss.Color("235")).
    Bold(true).
    Italic(true).
    Underline(true)

// Apply style
rendered := style.Render("Hello, World!")
```

### Color Types

```go
// 1. ANSI colors (0-255)
lipgloss.Color("1")     // Red
lipgloss.Color("15")    // White
lipgloss.Color("240")   // Gray

// 2. Hex colors
lipgloss.Color("#FF5733")

// 3. Adaptive colors (light/dark)
lipgloss.AdaptiveColor{
    Light: "#000000",  // Black in light mode
    Dark:  "#FFFFFF",  // White in dark mode
}

// 4. Named colors
lipgloss.Color("red")
lipgloss.Color("blue")
```

### Common Style Properties

```go
style := lipgloss.NewStyle().
    // Colors
    Foreground(lipgloss.Color("205")).
    Background(lipgloss.Color("235")).

    // Text styling
    Bold(true).
    Italic(true).
    Underline(true).
    Strikethrough(true).
    Blink(true).

    // Spacing
    Padding(1, 2, 1, 2).    // top, right, bottom, left
    Margin(1, 2).            // vertical, horizontal
    PaddingLeft(2).
    MarginTop(1).

    // Dimensions
    Width(50).
    Height(10).
    MaxWidth(80).
    MaxHeight(20).

    // Alignment
    Align(lipgloss.Center).
    AlignVertical(lipgloss.Center).

    // Borders
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63")).
    BorderBackground(lipgloss.Color("235")).
    BorderTop(true).
    BorderLeft(true)
```

### Layout Functions

```go
// Join vertically (stack)
lipgloss.JoinVertical(
    lipgloss.Left,   // Alignment
    "Line 1",
    "Line 2",
    "Line 3",
)

// Join horizontally (side by side)
lipgloss.JoinHorizontal(
    lipgloss.Top,    // Alignment
    "Column 1",
    "Column 2",
)

// Place (position in box)
lipgloss.Place(
    width,
    height,
    lipgloss.Center,  // horizontal
    lipgloss.Center,  // vertical
    "Centered text",
)

// Calculate width/height
width := lipgloss.Width("some text")
height := lipgloss.Height("line 1\nline 2")
```

---

## Theme System

### Theme Structure

```go
// internal/tui/theme/theme.go

type Theme struct {
    // Backgrounds
    SelectedBackground lipgloss.AdaptiveColor

    // Borders
    PrimaryBorder   lipgloss.AdaptiveColor
    SecondaryBorder lipgloss.AdaptiveColor
    FaintBorder     lipgloss.AdaptiveColor

    // Text
    PrimaryText   lipgloss.AdaptiveColor
    SecondaryText lipgloss.AdaptiveColor
    FaintText     lipgloss.AdaptiveColor
    InvertedText  lipgloss.AdaptiveColor
    SuccessText   lipgloss.AdaptiveColor
    WarningText   lipgloss.AdaptiveColor
    ErrorText     lipgloss.AdaptiveColor

    // Icons
    NewContributorIconColor lipgloss.AdaptiveColor
    ContributorIconColor    lipgloss.AdaptiveColor
    CollaboratorIconColor   lipgloss.AdaptiveColor
    MemberIconColor         lipgloss.AdaptiveColor
    OwnerIconColor          lipgloss.AdaptiveColor

    // Icon characters
    NewContributorIcon string
    ContributorIcon    string
    CollaboratorIcon   string
    MemberIcon         string
    OwnerIcon          string
}
```

### Default Theme

```go
var DefaultTheme = &Theme{
    // Borders
    PrimaryBorder:   lipgloss.AdaptiveColor{Light: "013", Dark: "008"},
    SecondaryBorder: lipgloss.AdaptiveColor{Light: "008", Dark: "007"},
    FaintBorder:     lipgloss.AdaptiveColor{Light: "254", Dark: "000"},

    // Backgrounds
    SelectedBackground: lipgloss.AdaptiveColor{Light: "006", Dark: "008"},

    // Text
    PrimaryText:   lipgloss.AdaptiveColor{Light: "000", Dark: "015"},
    SecondaryText: lipgloss.AdaptiveColor{Light: "244", Dark: "251"},
    FaintText:     lipgloss.AdaptiveColor{Light: "007", Dark: "245"},
    InvertedText:  lipgloss.AdaptiveColor{Light: "015", Dark: "236"},
    SuccessText:   lipgloss.AdaptiveColor{Light: "002", Dark: "002"},
    WarningText:   lipgloss.AdaptiveColor{Light: "003", Dark: "003"},
    ErrorText:     lipgloss.AdaptiveColor{Light: "001", Dark: "001"},

    // Icon colors
    NewContributorIconColor: lipgloss.AdaptiveColor{Light: "077", Dark: "077"},
    ContributorIconColor:    lipgloss.AdaptiveColor{Light: "075", Dark: "075"},
    CollaboratorIconColor:   lipgloss.AdaptiveColor{Light: "178", Dark: "178"},

    // Icon characters (Nerd Fonts)
    NewContributorIcon: "✨",
    ContributorIcon:    "👤",
    CollaboratorIcon:   "🤝",
    MemberIcon:         "👥",
    OwnerIcon:          "👑",
}
```

### Loading Theme from Config

```go
func ParseTheme(cfg *config.Config) Theme {
    theme := *DefaultTheme

    if cfg.Theme.Colors != nil {
        // Override colors from config
        if cfg.Theme.Colors.Inline.Text.Primary != "" {
            theme.PrimaryText = lipgloss.AdaptiveColor{
                Light: string(cfg.Theme.Colors.Inline.Text.Primary),
                Dark:  string(cfg.Theme.Colors.Inline.Text.Primary),
            }
        }
        // ... more overrides
    }

    if cfg.Theme.Icons != nil {
        // Override icons from config
        if cfg.Theme.Icons.Inline.Contributor != "" {
            theme.ContributorIcon = cfg.Theme.Icons.Inline.Contributor
        }
        // ... more overrides
    }

    return theme
}
```

---

## Styles Structure

### Centralized Styles

```go
// internal/tui/context/styles.go

type Styles struct {
    Section      SectionStyles
    Tabs         TabsStyles
    Sidebar      SidebarStyles
    Footer       FooterStyles
    ListViewPort ListViewPortStyles
    // ... more
}

func GetStyles(theme theme.Theme) *Styles {
    return &Styles{
        Section:      getSectionStyles(theme),
        Tabs:         getTabsStyles(theme),
        Sidebar:      getSidebarStyles(theme),
        Footer:       getFooterStyles(theme),
        ListViewPort: getListViewPortStyles(theme),
    }
}
```

### Section Styles

```go
type SectionStyles struct {
    ContainerStyle      lipgloss.Style
    EmptyStateStyle     lipgloss.Style
    TitleStyle          lipgloss.Style
    BorderStyle         lipgloss.Style
    KeyStyle            lipgloss.Style
}

func getSectionStyles(theme theme.Theme) SectionStyles {
    return SectionStyles{
        ContainerStyle: lipgloss.NewStyle().
            Padding(0, 1),

        EmptyStateStyle: lipgloss.NewStyle().
            Foreground(theme.FaintText).
            Align(lipgloss.Center),

        TitleStyle: lipgloss.NewStyle().
            Foreground(theme.PrimaryText).
            Bold(true),

        BorderStyle: lipgloss.NewStyle().
            Border(lipgloss.NormalBorder()).
            BorderForeground(theme.PrimaryBorder),

        KeyStyle: lipgloss.NewStyle().
            Foreground(theme.SuccessText).
            Bold(true),
    }
}
```

### Table Styles

```go
type TableStyles struct {
    HeaderStyle       lipgloss.Style
    RowStyle          lipgloss.Style
    SelectedRowStyle  lipgloss.Style
    CellStyle         lipgloss.Style
}

func getTableStyles(theme theme.Theme) TableStyles {
    return TableStyles{
        HeaderStyle: lipgloss.NewStyle().
            Foreground(theme.SecondaryText).
            Bold(true).
            BorderBottom(true).
            BorderStyle(lipgloss.NormalBorder()).
            BorderForeground(theme.FaintBorder),

        RowStyle: lipgloss.NewStyle().
            Foreground(theme.PrimaryText),

        SelectedRowStyle: lipgloss.NewStyle().
            Foreground(theme.PrimaryText).
            Background(theme.SelectedBackground).
            Bold(true),

        CellStyle: lipgloss.NewStyle().
            PaddingRight(2),
    }
}
```

### Sidebar Styles

```go
type SidebarStyles struct {
    ContainerStyle  lipgloss.Style
    TitleStyle      lipgloss.Style
    ContentStyle    lipgloss.Style
    TabActiveStyle  lipgloss.Style
    TabInactiveStyle lipgloss.Style
}

func getSidebarStyles(theme theme.Theme) SidebarStyles {
    return SidebarStyles{
        ContainerStyle: lipgloss.NewStyle().
            Border(lipgloss.NormalBorder()).
            BorderLeft(true).
            BorderForeground(theme.PrimaryBorder).
            Padding(1, 2),

        TitleStyle: lipgloss.NewStyle().
            Foreground(theme.PrimaryText).
            Bold(true).
            Underline(true),

        ContentStyle: lipgloss.NewStyle().
            Foreground(theme.PrimaryText),

        TabActiveStyle: lipgloss.NewStyle().
            Foreground(theme.PrimaryText).
            Background(theme.SelectedBackground).
            Bold(true).
            Padding(0, 2),

        TabInactiveStyle: lipgloss.NewStyle().
            Foreground(theme.SecondaryText).
            Padding(0, 2),
    }
}
```

---

## Adaptive Colors

### Light/Dark Mode Detection

```go
import "github.com/muesli/termenv"

func detectColorMode() bool {
    return termenv.HasDarkBackground()
}
```

### Using Adaptive Colors

```go
// Define adaptive color
borderColor := lipgloss.AdaptiveColor{
    Light: "#000000",  // Black border in light mode
    Dark:  "#FFFFFF",  // White border in dark mode
}

style := lipgloss.NewStyle().
    BorderForeground(borderColor)

// Lipgloss automatically picks correct color based on terminal
```

### Color Palette

```
Light Mode              Dark Mode
----------              ---------
Text:      #000000      Text:      #FFFFFF
Secondary: #666666      Secondary: #CCCCCC
Border:    #CCCCCC      Border:    #333333
Selected:  #E6E6E6      Selected:  #333333
Success:   #00AA00      Success:   #00FF00
Error:     #AA0000      Error:     #FF0000
```

---

## Layout and Spacing

### Responsive Width

```go
// Adapt to terminal width
func (m Model) View() string {
    availableWidth := m.ctx.ScreenWidth

    // Subtract sidebar if open
    if m.sidebar.IsOpen() {
        availableWidth -= sidebarWidth
    }

    // Create style with available width
    containerStyle := lipgloss.NewStyle().Width(availableWidth)

    return containerStyle.Render(content)
}
```

### Truncation

```go
// Truncate long text
func truncate(s string, maxWidth int) string {
    if lipgloss.Width(s) <= maxWidth {
        return s
    }
    return s[:maxWidth-3] + "..."
}

// Or use Lipgloss
style := lipgloss.NewStyle().
    Width(maxWidth).
    MaxWidth(maxWidth)

truncated := style.Render(longText)  // Auto-truncates
```

### Alignment

```go
// Center text in container
centered := lipgloss.Place(
    containerWidth,
    containerHeight,
    lipgloss.Center,
    lipgloss.Center,
    text,
)

// Left-align
leftAligned := lipgloss.NewStyle().
    Width(containerWidth).
    Align(lipgloss.Left).
    Render(text)

// Right-align
rightAligned := lipgloss.NewStyle().
    Width(containerWidth).
    Align(lipgloss.Right).
    Render(text)
```

### Borders

```go
// Border styles
lipgloss.RoundedBorder()  // ╭─╮
lipgloss.NormalBorder()   // ┌─┐
lipgloss.ThickBorder()    // ┏━┓
lipgloss.DoubleBorder()   // ╔═╗

// Partial borders
style := lipgloss.NewStyle().
    Border(lipgloss.NormalBorder()).
    BorderTop(true).
    BorderLeft(false).
    BorderRight(false).
    BorderBottom(true)

// Custom border
customBorder := lipgloss.Border{
    Top:         "─",
    Bottom:      "─",
    Left:        "│",
    Right:       "│",
    TopLeft:     "┌",
    TopRight:    "┐",
    BottomLeft:  "└",
    BottomRight: "┘",
}
```

---

## User Customization

### Config File Theming

Users can customize colors in `.gh-dash.yml`:

```yaml
theme:
  colors:
    text:
      primary: "#E0E0E0"
      secondary: "#808080"
    border:
      primary: "#0d0d0d"
    background:
      selected: "#2d2d2d"

  icons:
    contributor: "👤"
    owner: "👑"

  ui:
    sectionsShowCount: true
    table:
      showSeparator: true
      compact: false
```

### Loading Custom Colors

```go
// User provides hex color
type HexColor string

type ColorThemeText struct {
    Primary   HexColor `yaml:"primary"`
    Secondary HexColor `yaml:"secondary"`
    // ...
}

// Parse and apply
func ParseTheme(cfg *config.Config) Theme {
    theme := *DefaultTheme

    if cfg.Theme.Colors.Inline.Text.Primary != "" {
        theme.PrimaryText = lipgloss.AdaptiveColor{
            Light: string(cfg.Theme.Colors.Inline.Text.Primary),
            Dark:  string(cfg.Theme.Colors.Inline.Text.Primary),
        }
    }

    return theme
}
```

### Icon Customization

```yaml
theme:
  icons:
    newcontributor: "✨"
    contributor: "🔰"
    collaborator: "🤝"
    member: "👥"
    owner: "👑"
```

---

## Best Practices

### 1. Use Theme Colors

```go
// Good: Use theme colors
style := lipgloss.NewStyle().
    Foreground(m.ctx.Theme.PrimaryText).
    Background(m.ctx.Theme.SelectedBackground)

// Bad: Hardcode colors
style := lipgloss.NewStyle().
    Foreground(lipgloss.Color("255")).
    Background(lipgloss.Color("0"))
```

### 2. Reuse Styles

```go
// Good: Centralized styles
style := m.ctx.Styles.Section.TitleStyle
rendered := style.Render(title)

// Bad: Create styles everywhere
style := lipgloss.NewStyle().Bold(true).Render(title)
```

### 3. Responsive Design

```go
// Good: Adapt to screen size
width := min(m.ctx.ScreenWidth-10, 100)
style := lipgloss.NewStyle().Width(width)

// Bad: Fixed width
style := lipgloss.NewStyle().Width(80)  // Might overflow
```

### 4. Calculate Dimensions

```go
// Good: Calculate based on content
width := lipgloss.Width(content) + padding*2
style := lipgloss.NewStyle().Width(width)

// Bad: Guess width
style := lipgloss.NewStyle().Width(50)
```

### 5. Composition

```go
// Good: Compose styles
baseStyle := lipgloss.NewStyle().Padding(1, 2)
titleStyle := baseStyle.Copy().Bold(true)
errorStyle := baseStyle.Copy().Foreground(theme.ErrorText)

// Bad: Duplicate style definitions
titleStyle := lipgloss.NewStyle().Padding(1, 2).Bold(true)
errorStyle := lipgloss.NewStyle().Padding(1, 2).Foreground(theme.ErrorText)
```

### 6. Performance

```go
// Good: Cache styles
type Model struct {
    titleStyle lipgloss.Style  // Computed once
}

func NewModel() Model {
    return Model{
        titleStyle: lipgloss.NewStyle().Bold(true),
    }
}

// Bad: Compute every render
func (m Model) View() string {
    style := lipgloss.NewStyle().Bold(true)  // Computed every time
    return style.Render(m.title)
}
```

---

## Examples

### Styled Table Row

```go
func (m Model) renderRow(row RowData, isSelected bool) string {
    var style lipgloss.Style
    if isSelected {
        style = m.ctx.Styles.Table.SelectedRowStyle
    } else {
        style = m.ctx.Styles.Table.RowStyle
    }

    cells := []string{
        style.Render(fmt.Sprintf("%d", row.Number)),
        style.Render(truncate(row.Title, 50)),
        style.Render(row.Author),
    }

    return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}
```

### Styled Sidebar

```go
func (m Model) renderSidebar() string {
    title := m.ctx.Styles.Sidebar.TitleStyle.Render("Pull Request #123")

    content := m.ctx.Styles.Sidebar.ContentStyle.Render(
        "Description of the pull request...",
    )

    sidebar := lipgloss.JoinVertical(
        lipgloss.Left,
        title,
        "",
        content,
    )

    return m.ctx.Styles.Sidebar.ContainerStyle.Render(sidebar)
}
```

### Styled Footer

```go
func (m Model) renderFooter() string {
    left := m.ctx.Styles.Footer.LeftStyle.Render("Press ? for help")
    right := m.ctx.Styles.Footer.RightStyle.Render("Loading...")

    gap := m.ctx.ScreenWidth - lipgloss.Width(left) - lipgloss.Width(right)
    middle := strings.Repeat(" ", max(0, gap))

    return lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
}
```

---

## Summary

gh-dash styling is:

1. **Declarative**: Styles defined separately from content
2. **Themeable**: Colors customizable by users
3. **Adaptive**: Supports light/dark mode
4. **Responsive**: Adapts to terminal size
5. **Centralized**: Styles defined in one place
6. **Composable**: Complex styles built from simple ones

Key tools:
- **Lipgloss**: Styling and layout library
- **Theme system**: Color and icon management
- **Adaptive colors**: Light/dark mode support
- **Style composition**: Reusable style building blocks

When styling:
- Use theme colors (not hardcoded)
- Reuse centralized styles
- Calculate dimensions dynamically
- Test in light and dark mode
- Support terminal resize
- Keep styles simple and composable
