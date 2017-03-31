package hub

import "testing"

func TestHubListRemoveByIndex(t *testing.T) {
	hl := HubList{}

	hl = hl.add(newHub("1", nil))
	hl = hl.add(newHub("2", nil))
	hl = hl.add(newHub("3", nil))

	hl = hl.removeByIndex(len(hl) + 1)
	if len(hl) != 3 {
		t.Fatalf("result is %d, expect 3", len(hl))
	}
	hl = hl.removeByIndex(len(hl))
	if len(hl) != 3 {
		t.Fatalf("result is %d, expect 3", len(hl))
	}
	hl = hl.removeByIndex(len(hl) - 1)
	if len(hl) != 2 {
		t.Fatalf("result is %d, expect 2", len(hl))
	}
	hl = hl.removeByIndex(0)
	if len(hl) != 1 {
		t.Fatalf("result is %d, expect 1", len(hl))
	}
	b := hl[0].group == "2"
	if !b {
		t.Fatalf("result is %s, expect 2", hl[0].group)
	}
}
