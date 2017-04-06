package user

import (
	"testing"
)

func TestReplyContentListHead(t *testing.T) {
	list := replyContentList{
		[]byte("a"), []byte("b"),
	}

	head, tail := list.Head()
	if string(head) != "a" {
		t.Fatal("should be a")
	}
	if len(tail) != 1 {
		t.Fatal("should be 1")
	}
	if string(tail[0]) != "b" {
		t.Fatal("should be b")
	}
}
