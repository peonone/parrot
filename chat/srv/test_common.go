package srv

import "github.com/stretchr/testify/mock"

type mockStateStore struct {
	mock.Mock
}

func (s *mockStateStore) online(uid string, webNode string) error {
	returnVals := s.Called(uid, webNode)
	return returnVals.Error(0)
}

func (s *mockStateStore) offline(uid string) error {
	returnVals := s.Called(uid)
	return returnVals.Error(0)
}

func (s *mockStateStore) getOnlineWebNode(uid string) (string, error) {
	returnValues := s.Called(uid)
	return returnValues.String(0), returnValues.Error(1)
}

type mockMQSender struct {
	mock.Mock
}

func (s *mockMQSender) sendMQMsg(key string, body []byte) error {
	returnVals := s.Called(key, body)
	return returnVals.Error(0)
}
