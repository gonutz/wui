package wui

import "github.com/gonutz/w32"

// Color is a 24 bit color in BGR form. The alpha channel is always 0 and has no
// relevance.
type Color uint32

// R returns the red intensity in the Color, 0 means no red, 255 means full red.
func (c Color) R() uint8 { return uint8(c & 0xFF) }

// G returns the green intensity in the Color, 0 means no green, 255 means full
// green.
func (c Color) G() uint8 { return uint8((c & 0xFF00) >> 8) }

// B returns the blue intensity in the Color, 0 means no blue, 255 means full
// blue.
func (c Color) B() uint8 { return uint8((c & 0xFF0000) >> 16) }

// RBG creates a new Color with the given intensities. r,g,b means red, green,
// blue. Value 0 is dark, 255 is full intensity.
func RGB(r, g, b uint8) Color {
	return Color(r) + Color(g)<<8 + Color(b)<<16
}

// These are predefined colors defined by the current Windows theme.
var (
	// Scroll bar gray area.
	ColorScrollBar = sysColor(w32.COLOR_SCROLLBAR)

	// Desktop.
	ColorBackground = sysColor(w32.COLOR_BACKGROUND)

	// Desktop.
	ColorDesktop = sysColor(w32.COLOR_DESKTOP)

	// Active window title bar. The associated foreground color is
	// COLOR_CAPTIONTEXT. Specifies the left side color in the color gradient of
	// an active window's title bar if the gradient effect is enabled.
	ColorActiveCaption = sysColor(w32.COLOR_ACTIVECAPTION)

	// Inactive window caption. The associated foreground color is
	// COLOR_INACTIVECAPTIONTEXT. Specifies the left side color in the color
	// gradient of an inactive window's title bar if the gradient effect is
	// enabled.
	ColorInactiveCaption = sysColor(w32.COLOR_INACTIVECAPTION)

	// Menu background. The associated foreground color is COLOR_MENUTEXT.
	ColorMenu = sysColor(w32.COLOR_MENU)

	// Window background. The associated foreground colors are COLOR_WINDOWTEXT
	// and COLOR_HOTLITE.
	ColorWindow = sysColor(w32.COLOR_WINDOW)

	// Window frame.
	ColorWindowFrame = sysColor(w32.COLOR_WINDOWFRAME)

	// Text in menus. The associated background color is COLOR_MENU.
	ColorMenuText = sysColor(w32.COLOR_MENUTEXT)

	// Text in windows. The associated background color is COLOR_WINDOW.
	ColorWindowText = sysColor(w32.COLOR_WINDOWTEXT)

	// Text in caption, size box, and scroll bar arrow box. The associated
	// background color is COLOR_ACTIVECAPTION.
	ColorCaptionText = sysColor(w32.COLOR_CAPTIONTEXT)

	// Active window border.
	ColorActiveBorder = sysColor(w32.COLOR_ACTIVEBORDER)

	// Inactive window border.
	ColorInactiveBorder = sysColor(w32.COLOR_INACTIVEBORDER)

	// Background color of multiple document interface (MDI) applications.
	ColorAppWorkspace = sysColor(w32.COLOR_APPWORKSPACE)

	// Item(s) selected in a control. The associated foreground color is
	// COLOR_HIGHLIGHTTEXT.
	ColorHighlight = sysColor(w32.COLOR_HIGHLIGHT)

	// Text of item(s) selected in a control. The associated background color is
	// COLOR_HIGHLIGHT.
	ColorHighlightText = sysColor(w32.COLOR_HIGHLIGHTTEXT)

	// Face color for three-dimensional display elements and for dialog box
	// backgrounds.
	Color3DFace = sysColor(w32.COLOR_3DFACE)

	// Face color for three-dimensional display elements and for dialog box
	// backgrounds. The associated foreground color is COLOR_BTNTEXT.
	ColorButtonFace = sysColor(w32.COLOR_BTNFACE)

	// Shadow color for three-dimensional display elements (for edges facing
	// away from the light source).
	Color3DShadow = sysColor(w32.COLOR_3DSHADOW)

	// Shadow color for three-dimensional display elements (for edges facing
	// away from the light source).
	ColorButtonShadow = sysColor(w32.COLOR_BTNSHADOW)

	// Grayed (disabled) text. This color is set to 0 if the current display
	// driver does not support a solid gray color.
	ColorGrayText = sysColor(w32.COLOR_GRAYTEXT)

	// Text on push buttons. The associated background color is COLOR_BTNFACE.
	ColorButtonText = sysColor(w32.COLOR_BTNTEXT)

	// Color of text in an inactive caption. The associated background color is
	// COLOR_INACTIVECAPTION.
	ColorInactiveCaptionText = sysColor(w32.COLOR_INACTIVECAPTIONTEXT)

	// Highlight color for three-dimensional display elements (for edges facing
	// the light source.)
	Color3DHighlight = sysColor(w32.COLOR_3DHIGHLIGHT)

	// Highlight color for three-dimensional display elements (for edges facing
	// the light source.)
	ColorButtonHighlight = sysColor(w32.COLOR_BTNHIGHLIGHT)

	// Dark shadow for three-dimensional display elements.
	Color3DDarkShadow = sysColor(w32.COLOR_3DDKSHADOW)

	// Light color for three-dimensional display elements (for edges facing the
	// light source.)
	Color3DLight = sysColor(w32.COLOR_3DLIGHT)

	// Text color for tooltip controls. The associated background color is
	// COLOR_INFOBK.
	ColorInfoText = sysColor(w32.COLOR_INFOTEXT)

	// Background color for tooltip controls. The associated foreground color is
	// COLOR_INFOTEXT.
	ColorInfoBackground = sysColor(w32.COLOR_INFOBK)

	// Color for a hyperlink or hot-tracked item. The associated background
	// color is COLOR_WINDOW.
	ColorHotlight = sysColor(w32.COLOR_HOTLIGHT)

	// Right side color in the color gradient of an active window's title bar.
	// COLOR_ACTIVECAPTION specifies the left side color. Use
	// SPI_GETGRADIENTCAPTIONS with the SystemParametersInfo function to
	// determine whether the gradient effect is enabled.
	ColorGradientActiveCaption = sysColor(w32.COLOR_GRADIENTACTIVECAPTION)

	// Right side color in the color gradient of an inactive window's title bar.
	// COLOR_INACTIVECAPTION specifies the left side color.
	ColorGradientInactiveCaption = sysColor(w32.COLOR_GRADIENTINACTIVECAPTION)

	// The color used to highlight menu items when the menu appears as a flat
	// menu (see SystemParametersInfo). The highlighted menu item is outlined
	// with COLOR_HIGHLIGHT. Windows 2000: This value is not supported.
	ColorMenuHighlight = sysColor(w32.COLOR_MENUHILIGHT)

	// The background color for the menu bar when menus appear as flat menus
	// (see SystemParametersInfo). However, COLOR_MENU continues to specify the
	// background color of the menu popup. Windows 2000: This value is not
	// supported.
	ColorMenuBar = sysColor(w32.COLOR_MENUBAR)
)

func sysColor(index int) Color {
	return Color(w32.GetSysColor(index))
}
