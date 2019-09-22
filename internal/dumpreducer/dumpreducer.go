package dumpreducer

import (
	"container/heap"
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/negapedia/wikibrief"
	"github.com/negapedia/wikitfidf/internal/structures"
	"github.com/negapedia/wikitfidf/internal/utils"
)

// DumpReducer reduce the page information applying filters to it, like revert time frame, revert number and special page list
func DumpReducer(ctx context.Context, fail func(error) error, in <-chan wikibrief.EvolvingPage, resultDir string, nRevision int) {
	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := make([]structures.Revision, nRevision)
			for {
				select {
				case page, ok := <-in:
					if !ok {
						return
					}
					n := topN(buffer[:0], page.Revisions)
					writePage(page, buffer[:n], resultDir, fail)
				case <-ctx.Done():
					return
				}
			}

		}()
	}
	wg.Wait()
}

//topN is topN filter (based on a min-heap of structures.Revision with limited capacity).
func topN(top []structures.Revision, it <-chan wikibrief.Revision) (n int) {
	h := revisionHeap(top[:0])
	for r := range it {
		revision := structures.Revision{Timestamp: r.Timestamp, Text: r.Text}
		switch {
		case len(h) < cap(h): //There is space
			heap.Push(&h, revision)
		default: //len(h) == cap(h) : first is the youngest
			h[0] = revision
			heap.Fix(&h, 0)
		}
	}

	sort.Sort(h)

	return len(h)
}

// An revisionHeap is a min-heap of WeighedEdge.
type revisionHeap []structures.Revision

func (h revisionHeap) Len() int           { return len(h) }
func (h revisionHeap) Less(i, j int) bool { return h[i].Timestamp.Before(h[j].Timestamp) }
func (h revisionHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *revisionHeap) Push(x interface{}) {
	*h = append(*h, x.(structures.Revision))
}

func (h *revisionHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func writePage(page wikibrief.EvolvingPage, revArray []structures.Revision, resultDir string, fail func(error) error) {
	if len(revArray) == 0 {
		return
	}

	filename := filepath.Join(resultDir, strings.Repeat("0", 20-len(fmt.Sprint(page.PageID)))+fmt.Sprint(page.PageID)+".json")
	data := structures.Page{PageID: page.PageID, TopicID: page.TopicID, Revision: revArray}
	if err := utils.Write2JSON(filename, data); err != nil {
		fail(err)
	}
}
