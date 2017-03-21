package tests

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type userRefreshRequest struct {
	GroupID string
	count   int
}

type userRefreshRequestArray []*userRefreshRequest

func (this userRefreshRequestArray) Len() int           { return len(this) }
func (this userRefreshRequestArray) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }
func (this userRefreshRequestArray) Less(i, j int) bool { return this[i].count < this[j].count }

func TestSortArray(t *testing.T) {
	Convey("测试数组排序", t, func() {
		array := userRefreshRequestArray{}
		array = append(array, &userRefreshRequest{
			count: 1,
		})
		array = append(array, &userRefreshRequest{
			count: 0,
		})
		array = append(array, &userRefreshRequest{
			count: 2,
		})

		sort.Sort(array)

		So(array[0].count, ShouldEqual, 0)

	})
}

func TestSlice(t *testing.T) {
	Convey("测试 slice 截取", t, func() {
		ints := []int{1, 2, 3}
		s := ints[3:]
		So(len(s), ShouldEqual, 0)
	})
}

func getGoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
