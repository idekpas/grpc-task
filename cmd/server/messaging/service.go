package messaging

import (
	"log"
	"sync"
)

type Service struct {
	topics map[string]Topic
}

type Topic struct {
	subs map[string]chan string
}

func (t Topic) addSub(subID string) {
	_, ok := t.subs[subID]
	if !ok {
		t.subs[subID] = make(chan string)
	}
}

func NewService() Service {
	t := make(map[string]Topic)
	return Service{topics: t}
}

func NewTopic() Topic {
	m := make(map[string]chan string)
	return Topic{subs: m}
}

func (s Service) Subscribe(topicID string, subID string) {
	_, ok := s.topics[topicID]
	if !ok {
		s.topics[topicID] = NewTopic()
	}
	s.topics[topicID].addSub(subID)
}

func (s Service) Publish(topicID string, names ...string) {
	defer log.Printf("publishing complete")

	t, ok := s.topics[topicID]
	if !ok {
		return
	}

	var wg = &sync.WaitGroup{}
	for _, c := range t.subs {

		wg.Add(1)
		go func(ch chan<- string) {
			defer wg.Done()
			for _, name := range names {
				ch <- name
			}
		}(c)
	}
	wg.Wait()
}

func (s Service) Pull(topicID string) []string {
	defer log.Printf("pulling complete")

	t, ok := s.topics[topicID]
	if !ok {
		return make([]string, 0)
	}

	names := make([]string, 0)

	var wg = &sync.WaitGroup{}
	for _, c := range t.subs {
		wg.Add(1)
		go func(ch <-chan string) {
			defer wg.Done()
			for v := range ch {
				names = append(names, v)
				log.Printf("names: %v\n", names)
			}
		}(c)
	}
	wg.Wait()
	return names
}
