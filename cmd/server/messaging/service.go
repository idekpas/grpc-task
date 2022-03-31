package messaging

type Service struct {
	data map[string]Data
}

type Data struct {
	q  []string
	ch chan string
}

func NewService() Service {
	d := make(map[string]Data)
	return Service{data: d}
}

func (s Service) addData(id string) {
	q := make([]string, 0)
	ch := make(chan string)
	d := Data{q: q, ch: ch}

	s.data[id] = d
}

func (s Service) Subscribe(id string) chan string {
	if d, ok := s.data[id]; !ok {
		s.addData(id)
		return s.Subscribe(id)
	} else {
		return d.ch
	}
}

func (s Service) Publish(ch chan string, names ...string) {
	for _, v := range names {
		ch <- v
	}
}

func (s Service) Pull(ch chan string) string {
	return <-ch
}
