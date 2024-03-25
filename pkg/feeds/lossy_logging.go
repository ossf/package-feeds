package feeds

import (
	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/pkg/events"
)

type LossyFeedAlerter struct {
	previousPackages map[string][]*Package
	eventHandler     *events.Handler
}

// Creates a LossyFeedAlerter, capable of tracking packages and identifying
// potential loss in feeds using RSS style APIs. This can only be used in
// feeds which produce an overlap of packages upon their requests to the API,
// if a timestamp is used to query the API then loss is unlikely due to requesting
// data since a previous query.
func NewLossyFeedAlerter(eventHandler *events.Handler) *LossyFeedAlerter {
	return &LossyFeedAlerter{
		eventHandler:     eventHandler,
		previousPackages: map[string][]*Package{},
	}
}

// Processes a new collection of packages and compares against the previously processed
// slice of packages, if an overlap is not detected this is a sign of potential loss of
// data and the configured event handler is notified via a LossyFeedEvent.
func (lfa *LossyFeedAlerter) ProcessPackages(feed string, packages []*Package) {
	pkgs := make([]*Package, len(packages))
	copy(pkgs, packages)

	SortPackages(pkgs)

	previousPackages, ok := lfa.previousPackages[feed]
	nonZeroResults := len(pkgs) > 0 && len(previousPackages) > 0
	if ok && nonZeroResults {
		if !findOverlap(pkgs, previousPackages) {
			err := lfa.eventHandler.DispatchEvent(events.LossyFeedEvent{
				Feed: feed,
			})
			if err != nil {
				log.WithError(err).Error("failed to dispatch event via event handler")
			}
		}
	}
	lfa.previousPackages[feed] = pkgs
}

// Checks whether there is an overlap in package creation date between a result
// and a previous result. This assumes that pollResult.packages is sorted by
// CreatedDate in order of most recent first.
func findOverlap(latestPackages, previousPackages []*Package) bool {
	rOldestPkg := latestPackages[len(latestPackages)-1]
	previousResultMostRecent := previousPackages[0]

	afterDate := previousResultMostRecent.CreatedDate.After(rOldestPkg.CreatedDate)
	equalDate := previousResultMostRecent.CreatedDate.Equal(rOldestPkg.CreatedDate)
	return afterDate || equalDate
}
