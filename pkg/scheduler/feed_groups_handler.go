package scheduler

import (
	"fmt"
	"net/http"
	"strings"
)

type FeedGroupsHandler struct {
	feedGroups []*FeedGroup
}

func NewFeedGroupsHandler(feeds []*FeedGroup) *FeedGroupsHandler {
	return &FeedGroupsHandler{feedGroups: feeds}
}

func (srv *FeedGroupsHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	resultChannel := make(chan groupResult, len(srv.feedGroups))
	numPublished := 0
	var pollErr, pubErr error
	var errStrings []string
	for _, group := range srv.feedGroups {
		go func(group *FeedGroup) {
			result := group.pollAndPublish()
			resultChannel <- result
		}(group)
	}
	for range srv.feedGroups {
		result := <-resultChannel
		numPublished += result.numPublished
		if result.pollErr != nil {
			pollErr = result.pollErr
		}
		if result.pubErr != nil {
			pubErr = result.pubErr
		}
	}
	for _, err := range []error{pollErr, pubErr} {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	if len(errStrings) > 0 {
		http.Error(w, strings.Join(errStrings, "\n")+fmt.Sprintf("\n%d packages successfully processed, see log for details",
			numPublished),
			http.StatusInternalServerError)
		return
	}
	_, err := w.Write([]byte(fmt.Sprintf("%d packages processed", numPublished)))
	if err != nil {
		http.Error(w, "unexpected error during http server write: %w", http.StatusInternalServerError)
	}
}
