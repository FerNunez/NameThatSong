package guesser

import "fmt"

type MusicPlayer struct {
	Queue   []string
	Current string
}

func NewMusicPlayer() MusicPlayer {
	return MusicPlayer{}
}

func (mp *MusicPlayer) Next() bool {
	if len(mp.Queue) > 0 {




		mp.Current = mp.Queue[0]
		mp.Queue = mp.Queue[1:]
		fmt.Println("Current goign song; ", mp.Current)
		return true
	}

	fmt.Println("Empty QUEUE ")
	return false
}

//func (mp *MusicPlayer) addQueue(songId int) {
//	mp.queue = append(mp.queue, songId)
//}
//
//func (mp *MusicPlayer) popQueue() int {
//	if len(mp.queue) >= 1 {
//		id := mp.queue[0]
//		mp.queue = mp.queue[1:len(mp.queue)]
//		return id
//	}
//	return 0
//}
