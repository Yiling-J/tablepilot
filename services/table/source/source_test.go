package source

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSource_Indexer(t *testing.T) {
	indexer := newIndexer(false, false, 5)
	nums := []int{}
	for i := 0; i < 10; i++ {
		nums = append(nums, indexer.nextIndex())
	}
	require.Equal(t, []int{0, 1, 2, 3, 4, 0, 1, 2, 3, 4}, nums)

	indexer = newIndexer(true, false, 5)
	numsCounter := map[int]int{}
	nums = []int{}
	for i := 0; i < 50; i++ {
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

	indexer = newIndexer(true, true, 5)
	nums = []int{}
	for i := 0; i < 10; i++ {
		nums = append(nums, indexer.nextIndex())
		if (i+1)%5 == 0 {
			require.NotEqual(t, []int{0, 1, 2, 3, 4}, nums)
			require.ElementsMatch(t, []int{0, 1, 2, 3, 4}, nums)
			nums = []int{}
		}
	}
}
