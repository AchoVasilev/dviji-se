package templates

// AdVariant selects where a slot may appear and how big it is. The variant is
// only ever a class name: the .ad-slot-* rules in main.css own the sizing and
// the breakpoints, so a new placement is a variant plus a rule, not another
// copy of the markup.
type AdVariant string

const (
	// AdRail is the column beside the content, shown only where there is room
	// for it next to the page.
	AdRail AdVariant = "rail"

	// AdBanner is the narrow screen placement: full width, one viewport tall,
	// scrolled past to reach the content.
	AdBanner AdVariant = "banner"

	// AdInline sits in the flow of a page at any width.
	AdInline AdVariant = "inline"
)

// Class is the variant's marker class.
func (v AdVariant) Class() string {
	return "ad-slot-" + string(v)
}
