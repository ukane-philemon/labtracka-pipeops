package patient

func (s *Server) doInBackground(action string, fn func() error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := fn()
		if err != nil {
			s.logger.Error("%s: %v", action, err)
		}
	}()
}
