package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"context"

	"github.com/negapedia/wikibrief"
)

//Filter applies specified filter to the pages channel and return it. If nothing is to be done, it just return it.
func Filter(ctx context.Context, fail func(error) error, in <-chan wikibrief.EvolvingPage, startDateFlag, endDateFlag, specialPageListFlag string) <-chan wikibrief.EvolvingPage {
	if startDateFlag+endDateFlag+specialPageListFlag == "" {
		return in
	}

	closedCh := make(chan wikibrief.EvolvingPage)
	close(closedCh)

	specialPages := map[string]bool{}
	for _, title := range strings.Split(specialPageListFlag, "-") {
		specialPages[title] = true
	}

	filterPage := func(p wikibrief.EvolvingPage) (isOk bool) {
		return len(specialPages) == 0 || specialPages[p.Title]
	}

	startDate, errStart := time.Parse(startDateFlag, "2019-01-01T15:00")
	endDate, errEnd := time.Parse(endDateFlag, "2019-01-01T15:00")
	switch {
	case startDateFlag == "":
		//startDateFlag already initialized to zero
	case errStart != nil:
		fail(fmt.Errorf("Error while parsing start date flag %v", startDateFlag))
		return closedCh
	case endDateFlag == "":
		endDate = time.Now().Add(24 * time.Hour)
	case errEnd != nil:
		fail(fmt.Errorf("Error while parsing end date flag %v", endDateFlag))
		return closedCh
	}

	filterRevision := func(r wikibrief.Revision) (isOk bool) {
		return r.Timestamp.After(startDate) && r.Timestamp.Before(endDate)
	}

	outCh := make(chan wikibrief.EvolvingPage, 50)

	go filter(ctx, fail, in, outCh, filterPage, filterRevision)

	return outCh
}

func filter(ctx context.Context, fail func(error) error, in <-chan wikibrief.EvolvingPage, out chan<- wikibrief.EvolvingPage,
	filterPage func(p wikibrief.EvolvingPage) bool, filterRevision func(r wikibrief.Revision) bool) {
	defer func() {
		close(out)
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range in {
				if !filterPage(page) {
					for range page.Revisions {
						//skip
					}
					continue
				}

				outRevCh := make(chan wikibrief.Revision, 200)
				newpage := page
				newpage.Revisions = outRevCh
				select {
				case out <- newpage:
					//Go on
				case <-ctx.Done():
					return
				}

				for rev := range page.Revisions {
					if !filterRevision(rev) {
						continue
					}
					select {
					case outRevCh <- rev:
						//Go on
					case <-ctx.Done():
						return
					}
				}
				close(outRevCh)
			}
		}()
	}
	wg.Wait()
}
