package source

import (
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/stretchr/testify/require"
)

func TestSource_Indexer(t *testing.T) {
	so := &ListSource{
		BasicSource: BasicSource{Type: "list"},
		Options:     []string{"a", "b", "c", "d", "e"},
	}
	indexer := NewIndexer(so, &ent.TableColumn{
		Random: false,
	})
	nums := []int{}
	for range 10 {
		nums = append(nums, indexer.nextIndex())
	}
	require.Equal(t, []int{0, 1, 2, 3, 4, 0, 1, 2, 3, 4}, nums)

	indexer = NewIndexer(so, &ent.TableColumn{
		Random:      true,
		Replacement: true,
	})
	numsCounter := map[int]int{}
	nums = []int{}
	for i := range 50 {
		j := indexer.nextIndex()
		nums = append(nums, j)
		numsCounter[j] += 1
		if (i+1)%5 == 0 {
			require.NotEqual(t, []int{0, 1, 2, 3, 4}, nums)
			nums = []int{}
		}
	}
	lt := 0
	gt := 0
	eq := 0
	for _, v := range numsCounter {
		switch {
		case v < 10:
			lt += 1
		case v == 10:
			eq += 1
		case v > 10:
			gt += 1
		}
	}
	require.Equal(t, 5, lt+eq+gt)
	require.True(t, lt > 0)
	require.True(t, gt > 0)
	require.True(t, eq >= 0)

	indexer = NewIndexer(so, &ent.TableColumn{
		Random: true,
	})
	nums = []int{}
	for i := range 10 {
		nums = append(nums, indexer.nextIndex())
		if (i+1)%5 == 0 {
			require.NotEqual(t, []int{0, 1, 2, 3, 4}, nums)
			require.ElementsMatch(t, []int{0, 1, 2, 3, 4}, nums)
			nums = []int{}
		}
	}

	indexer = NewIndexer(so, &ent.TableColumn{
		Random: false,
		Repeat: 2,
	})
	nums = []int{}
	for range 10 {
		nums = append(nums, indexer.nextIndex())
	}
	require.Equal(t, []int{0, 0, 1, 1, 2, 2, 3, 3, 4, 4}, nums)
}
