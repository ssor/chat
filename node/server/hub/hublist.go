package hub

// List is an array of hub
type List []*Hub

func (hl List) add(hs ...*Hub) List {
	hl = append(hl, hs...)
	return hl
}

// Contains return true if a hub already exists
func (hl List) Contains(id string) bool {
	for _, hub := range hl {
		if hub.group == id {
			return true
		}
	}
	return false
}

// Find return nil if no hub found
func (hl List) Find(id string) *Hub {
	for _, hub := range hl {
		if hub.group == id {
			return hub
		}
	}
	return nil
}

func (hl List) remove(id string) (list List, hub *Hub) {
	for index, h := range hl {
		if h.group == id {
			hub = h
			list = hl.removeByIndex(index)
			break
		}
	}
	return
}

func (hl List) removeByIndex(index int) (list List) {
	if index >= len(hl) { // out of range
		list = hl[:]
	} else {
		list = append(hl[:index], hl[index+1:]...)
	}
	return
}
