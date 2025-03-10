package guesser

type MusicPlayer struct {
	queue []int
}

func NewMusicPlayer() MusicPlayer {
	return MusicPlayer{}
}

func (mp *MusicPlayer) addQueue(songId int) {
	mp.queue = append(mp.queue, songId)
}

func (mp *MusicPlayer) popQueue() int {
	if len(mp.queue) >= 1 {
		id := mp.queue[0]
		mp.queue = mp.queue[1:len(mp.queue)]
		return id
	}
	return 0
}
