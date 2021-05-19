package scheduler

import (
	"fmt"
	"net/http"
)

type FeedGroupsHandler struct {
	feedGroups []*FeedGroup
}

type publishResult struct {
	numPublished int
	err          error
}

func NewFeedGroupsHandler(feeds []*FeedGroup) *FeedGroupsHandler {
	return &FeedGroupsHandler{feedGroups: feeds}
}

func (srv *FeedGroupsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	resultChannel := make(chan publishResult, len(srv.feedGroups))
	numPublished := 0
	for _, group := range srv.feedGroups {
		go func(group *FeedGroup) {
			numPublished, err := group.PollAndPublish()
			resultChannel <- publishResult{numPublished, err}
		}(group)
	}
	for range srv.feedGroups {
		result := <-resultChannel

		numPublished += result.numPublished
		if result.err != nil {
			http.Error(w, result.err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf("%d packages processed", numPublished)))
	if err != nil {
		http.Error(w, "unexpected error during http server write: %w", http.StatusInternalServerError)
	}
}
