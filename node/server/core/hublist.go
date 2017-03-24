package core

// HubList is an array of hub
type HubList []*Hub

func (hl HubList) add(hs ...*Hub) HubList {
	hl = append(hl, hs...)
	return hl
}

// Contains return true if a hub already exists
func (hl HubList) Contains(id string) bool {
	for _, hub := range hl {
		if hub.group == id {
			return true
		}
	}
	return false
}

// Find return nil if no hub found
func (hl HubList) Find(id string) *Hub {
	for _, hub := range hl {
		if hub.group == id {
			return hub
		}
	}
	return nil
}

func (hl HubList) remove(id string) (list HubList, hub *Hub) {
	for index, h := range hl {
		if h.group == id {
			hub = h
			list = hl.removeByIndex(index)
			break
		}
	}
	return
}

func (hl HubList) removeByIndex(index int) (list HubList) {
	if (index + 1) >= len(hl) { // out of range
		list = hl[:index]
	} else {
		list = append(hl[:index], hl[index+1:]...)
	}
	return
}
