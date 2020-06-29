package ring

import "github.com/bwmarrin/discordgo"

// Ring of messages for cache
type Ring struct {
	end    int
	cycled bool
	Buffer [1000]*discordgo.Message
}

// Push messages
func (r *Ring) Push(m []*discordgo.Message) {
	n := len(m)
	if n > 1000 {
		m = m[n-r.end:]
	}
	if !r.cycled && r.end+n > 999 {
		r.cycled = true
	}
	// Copy
	for _, mess := range m {
		r.Buffer[r.end] = mess
		r.end = (r.end + 1) % 1000
	}
}

// Get messages
func (r *Ring) Get(i int) *discordgo.Message {
	return r.Buffer[i]
}

// GetLast message
func (r *Ring) GetLast() *discordgo.Message {
	if r.end == 0 {
		if r.cycled {
			return r.Buffer[999]
		}
		return nil
	}
	return r.Buffer[r.end-1]
}

// GetFirst message still left in buffer
func (r *Ring) GetFirst() *discordgo.Message {
	if r.cycled {
		return r.Buffer[r.end]
	}
	return r.Buffer[0]
}

// Len of buf
func (r *Ring) Len() int {
	if r.cycled {
		return 1000
	}
	return r.end
}
